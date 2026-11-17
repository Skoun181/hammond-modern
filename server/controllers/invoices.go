package controllers

import (
	"github.com/akhilrex/hammond/db"
	"github.com/akhilrex/hammond/service"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"os"
	"strings"
)

func RegisterInvoiceController(router *gin.RouterGroup) {
	router.POST("/invoices", func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 10*1024*1024+1024*1024)
		file, err := c.FormFile("file")
		if err != nil || file.Size <= 0 || file.Size > 10*1024*1024 {
			c.JSON(400, gin.H{"error": "Rechnung muss höchstens 10 MB groß sein"})
			return
		}
		input, err := file.Open()
		if err != nil {
			c.JSON(400, gin.H{"error": "Datei konnte nicht gelesen werden"})
			return
		}
		head := make([]byte, 512)
		n, _ := input.Read(head)
		input.Close()
		kind := http.DetectContentType(head[:n])
		if kind != "application/pdf" && kind != "image/jpeg" && kind != "image/png" {
			c.JSON(400, gin.H{"error": "Nur PDF, JPG oder PNG erlaubt"})
			return
		}
		ext := ".pdf"
		if kind == "image/jpeg" {
			ext = ".jpg"
		}
		if kind == "image/png" {
			ext = ".png"
		}
		filepath := service.GetFilePath("rechnung" + ext)
		if err := c.SaveUploadedFile(file, filepath); err != nil {
			c.JSON(500, gin.H{"error": "Upload fehlgeschlagen"})
			return
		}
		item, err := service.CreateAttachment(filepath, file.Filename, file.Size, kind, c.MustGet("userId").(string))
		if err != nil {
			os.Remove(filepath)
			c.JSON(500, gin.H{"error": "Anhang konnte nicht gespeichert werden"})
			return
		}
		c.JSON(201, gin.H{"id": item.ID, "originalName": item.OriginalName})
	})
	router.GET("/invoices/:id/file", func(c *gin.Context) {
		var obj db.Attachment
		if err := db.DB.Where("id=? AND user_id=?", c.Param("id"), c.MustGet("userId").(string)).First(&obj).Error; err != nil {
			c.Status(404)
			return
		}
		var count int64
		db.DB.Model(&db.Expense{}).Where("invoice_attachment_id=?", obj.ID).Count(&count)
		if count == 0 {
			c.Status(404)
			return
		}
		// Content-Disposition prevents inline execution for a file that passes the upload checks.
		safe := strings.NewReplacer("\r", "", "\n", "", "\"", "").Replace(obj.OriginalName)
		c.Header("Content-Disposition", "attachment; filename=\""+safe+"\"")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("Content-Type", "application/octet-stream")
		file, err := os.Open(obj.Path)
		if err != nil {
			c.Status(404)
			return
		}
		defer file.Close()
		c.Status(200)
		io.Copy(c.Writer, file)
	})
}

func validInvoiceForUser(id string, user string) bool {
	if id == "" {
		return true
	}
	var count int64
	return db.DB.Model(&db.Attachment{}).Where("id=? AND user_id=?", id, user).Count(&count).Error == nil && count == 1
}
