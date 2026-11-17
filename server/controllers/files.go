package controllers

import (
	"io/ioutil"
	"net/http"
	"os"

	"github.com/akhilrex/hammond/db"
	"github.com/akhilrex/hammond/models"
	"github.com/akhilrex/hammond/service"
	"github.com/gin-gonic/gin"
)

func RegisterFilesController(router *gin.RouterGroup) {
	router.POST("/upload", uploadFile)

	router.GET("/attachments/:id/file", getAttachmentFile)
}

func uploadFile(c *gin.Context) {
	attachment, err := saveMultipleUploadedFile(c, "file")
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
	} else {
		c.JSON(http.StatusOK, attachment)
	}
}
func getAttachmentFile(c *gin.Context) {
	var searchByIdQuery models.SearchByIdQuery

	if c.ShouldBindUri(&searchByIdQuery) == nil {

		attachment, err := db.GetAttachmentById(searchByIdQuery.Id)
		if err == nil {
			if _, err = os.Stat(attachment.Path); os.IsNotExist(err) {
				c.Status(404)
			} else {
				c.File(attachment.Path)
			}
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
	}
}

func getFileBytes(c *gin.Context, fileVariable string) ([]byte, error) {
	if fileVariable == "" {
		fileVariable = "file"
	}
	formFile, err := c.FormFile(fileVariable)
	if err != nil {
		return nil, err
	}
	openedFile, _ := formFile.Open()
	return ioutil.ReadAll(openedFile)
}

func saveUploadedFile(c *gin.Context, fileVariable string) (*db.Attachment, error) {
	if fileVariable == "" {
		fileVariable = "file"
	}
	file, err := c.FormFile(fileVariable)
	if err != nil {
		return nil, err
	}
	filePath := service.GetFilePath(file.Filename)
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		return nil, err
	}

	return service.CreateAttachment(filePath, file.Filename, file.Size, file.Header.Get("Content-Type"), c.MustGet("userId").(string))
}
func saveMultipleUploadedFile(c *gin.Context, fileVariable string) ([]*db.Attachment, error) {
	if fileVariable == "" {
		fileVariable = "files"
	}
	form, err := c.MultipartForm()
	if err != nil {
		return nil, err
	}
	files := form.File[fileVariable]
	var toReturn []*db.Attachment
	for _, file := range files {
		filePath := service.GetFilePath(file.Filename)
		if err := c.SaveUploadedFile(file, filePath); err != nil {
			return nil, err
		}
		attachment, err := service.CreateAttachment(filePath, file.Filename, file.Size, file.Header.Get("Content-Type"), c.MustGet("userId").(string))
		if err != nil {
			return nil, err
		}

		toReturn = append(toReturn, attachment)
	}
	return toReturn, nil
}
