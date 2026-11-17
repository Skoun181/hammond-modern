package controllers

import (
	"github.com/akhilrex/hammond/db"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"time"
)

type driverRequest struct {
	Name          string     `json:"name" binding:"required"`
	LicenseExpiry *time.Time `json:"licenseExpiry"`
	Notes         string     `json:"notes"`
}

func RegisterDriverController(r *gin.RouterGroup) {
	r.GET("/drivers", func(c *gin.Context) {
		userID := c.MustGet("userId").(string)
		result := []db.Driver{}
		if err := db.DB.Where("owner_user_id = ?", userID).Order("name asc").Find(&result).Error; err != nil {
			c.JSON(500, gin.H{"error": "Fahrer konnten nicht geladen werden"})
			return
		}
		c.JSON(200, result)
	})
	r.POST("/drivers", func(c *gin.Context) {
		var input driverRequest
		if err := c.ShouldBindJSON(&input); err != nil || strings.TrimSpace(input.Name) == "" {
			c.JSON(422, gin.H{"error": "Name fehlt"})
			return
		}
		if len(input.Name) > 150 || len(input.Notes) > 2000 {
			c.JSON(422, gin.H{"error": "Eingabe zu lang"})
			return
		}
		obj := db.Driver{OwnerUserID: c.MustGet("userId").(string), Name: strings.TrimSpace(input.Name), LicenseExpiry: input.LicenseExpiry, Notes: input.Notes}
		if err := db.DB.Create(&obj).Error; err != nil {
			c.JSON(500, gin.H{"error": "Speichern fehlgeschlagen"})
			return
		}
		c.JSON(http.StatusCreated, obj)
	})
	r.PUT("/drivers/:id", func(c *gin.Context) {
		var input driverRequest
		if err := c.ShouldBindJSON(&input); err != nil || strings.TrimSpace(input.Name) == "" {
			c.JSON(422, gin.H{"error": "Name fehlt"})
			return
		}
		if len(input.Name) > 150 || len(input.Notes) > 2000 {
			c.JSON(422, gin.H{"error": "Eingabe zu lang"})
			return
		}
		var obj db.Driver
		if err := db.DB.Where("id = ? AND owner_user_id = ?", c.Param("id"), c.MustGet("userId").(string)).First(&obj).Error; err != nil {
			c.JSON(404, gin.H{"error": "Fahrer nicht gefunden"})
			return
		}
		obj.Name = strings.TrimSpace(input.Name)
		obj.Notes = input.Notes
		obj.LicenseExpiry = input.LicenseExpiry
		if err := db.DB.Save(&obj).Error; err != nil {
			c.JSON(500, gin.H{"error": "Speichern fehlgeschlagen"})
			return
		}
		c.JSON(200, obj)
	})
	r.DELETE("/drivers/:id", func(c *gin.Context) {
		result := db.DB.Where("id = ? AND owner_user_id = ?", c.Param("id"), c.MustGet("userId").(string)).Delete(&db.Driver{})
		if result.Error != nil {
			c.JSON(500, gin.H{"error": "Löschen fehlgeschlagen"})
			return
		}
		if result.RowsAffected == 0 {
			c.JSON(404, gin.H{"error": "Fahrer nicht gefunden"})
			return
		}
		c.Status(http.StatusNoContent)
	})
}
