package controllers

import (
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/akhilrex/hammond/db"
	"github.com/akhilrex/hammond/service"
	"github.com/gin-gonic/gin"
)

const maxAttachment = 10 * 1024 * 1024

func RegisterFilesController(router *gin.RouterGroup) {
	router.POST("/upload", uploadFile)
	router.GET("/attachments/:id/file", getAttachmentFile)
}

func uploadFile(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 5*maxAttachment+1024*1024)
	form, err := c.MultipartForm()
	if err != nil || len(form.File["file"]) == 0 || len(form.File["file"]) > 5 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ein bis fünf Anhänge erforderlich"})
		return
	}
	files := form.File["file"]
	types := make([]string, len(files))
	for i, file := range files {
		types[i], err = attachmentType(file)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}
	attachments := make([]*db.Attachment, 0, len(files))
	for i, file := range files {
		attachment, err := storeAttachment(c, file, types[i])
		if err != nil {
			for _, item := range attachments {
				if db.DB.Unscoped().Delete(item).Error == nil {
					_ = os.Remove(item.Path)
				}
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Upload fehlgeschlagen"})
			return
		}
		attachments = append(attachments, attachment)
	}
	c.JSON(http.StatusOK, attachments)
}

func getAttachmentFile(c *gin.Context) {
	user := c.GetString("userId")
	var attachment db.Attachment
	shared := db.DB.Table("vehicle_attachments AS va").Select("va.attachment_id").
		Joins("JOIN user_vehicles AS uv ON uv.vehicle_id = va.vehicle_id").
		Joins("JOIN vehicles AS v ON v.id = va.vehicle_id").
		Where("uv.user_id = ? AND va.deleted_at IS NULL AND uv.deleted_at IS NULL AND v.deleted_at IS NULL", user)
	err := db.DB.Where("id = ? AND (user_id = ? OR id IN (?))", c.Param("id"), user, shared).First(&attachment).Error
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	sendAttachment(c, &attachment)
}

func sendAttachment(c *gin.Context, attachment *db.Attachment) {
	root, err := os.OpenRoot(os.Getenv("DATA"))
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	defer root.Close()
	file, err := root.Open(filepath.Base(attachment.Path))
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() {
		c.Status(http.StatusNotFound)
		return
	}
	c.Header("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": attachment.OriginalName}))
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("Content-Type", "application/octet-stream")
	http.ServeContent(c.Writer, c.Request, "", info.ModTime(), file)
}

func getFileBytes(c *gin.Context, name string) ([]byte, error) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxAttachment+1024*1024)
	formFile, err := c.FormFile(name)
	if err != nil {
		return nil, err
	}
	if formFile.Size > maxAttachment {
		return nil, errors.New("Datei ist zu groß")
	}
	file, err := formFile.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return io.ReadAll(file)
}

func attachmentType(file *multipart.FileHeader) (string, error) {
	if file.Size <= 0 || file.Size > maxAttachment {
		return "", errors.New("Anhang muss zwischen 1 Byte und 10 MiB groß sein")
	}
	source, err := file.Open()
	if err != nil {
		return "", err
	}
	defer source.Close()
	var header [512]byte
	n, err := io.ReadFull(source, header[:])
	if err != nil && err != io.ErrUnexpectedEOF {
		return "", err
	}
	kind := http.DetectContentType(header[:n])
	switch kind {
	case "application/pdf", "image/jpeg", "image/png", "image/webp":
		return kind, nil
	}
	return "", errors.New("Nur PDF, JPEG, PNG oder WebP erlaubt")
}

func storeAttachment(c *gin.Context, file *multipart.FileHeader, kind string) (*db.Attachment, error) {
	extensions := map[string]string{"application/pdf": ".pdf", "image/jpeg": ".jpg", "image/png": ".png", "image/webp": ".webp"}
	path := service.GetFilePath("upload" + extensions[kind])
	if err := c.SaveUploadedFile(file, path, 0600); err != nil {
		_ = os.Remove(path)
		return nil, err
	}
	attachment, err := service.CreateAttachment(path, filepath.Base(file.Filename), file.Size, kind, c.GetString("userId"))
	if err != nil {
		_ = os.Remove(path)
	}
	return attachment, err
}
