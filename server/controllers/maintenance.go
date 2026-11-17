package controllers

import (
	"net/http"
	"strings"
	"time"

	"github.com/akhilrex/hammond/db"
	"github.com/gin-gonic/gin"
)

type maintenanceInput struct {
	VehicleID string     `json:"vehicleId" binding:"required"`
	Title     string     `json:"title" binding:"required"`
	Notes     string     `json:"notes"`
	DueDate   *time.Time `json:"dueDate"`
	DueOdo    *int       `json:"dueOdo"`
	Completed bool       `json:"completed"`
}

func canAccessMaintenanceVehicle(userID, vehicleID string) bool {
	var count int64
	if vehicleID == "" { return false }
	return db.DB.Model(&db.UserVehicle{}).
		Where("user_id = ? AND vehicle_id = ?", userID, vehicleID).
		Count(&count).Error == nil && count > 0
}

func RegisterMaintenanceController(router *gin.RouterGroup) {
	router.GET("/maintenance", func(c *gin.Context) {
		userID := c.MustGet("userId").(string)
		var tasks []db.MaintenanceTask
		err := db.DB.Where("owner_user_id = ?", userID).Order("created_at DESC").Find(&tasks).Error
		if err != nil { c.JSON(500, gin.H{"error": "Wartungen konnten nicht geladen werden"}); return }
		if tasks == nil { tasks = []db.MaintenanceTask{} }
		c.JSON(200, tasks)
	})
	router.POST("/maintenance", func(c *gin.Context) {
		var input maintenanceInput
		if c.ShouldBindJSON(&input) != nil || strings.TrimSpace(input.Title) == "" || len(input.Title) > 120 || len(input.Notes) > 2000 || (input.DueOdo == nil && input.DueDate == nil) || (input.DueOdo != nil && *input.DueOdo < 0) {
			c.JSON(422, gin.H{"error": "Titel sowie Datum oder Kilometerstand angeben"}); return
		}
		userID := c.MustGet("userId").(string)
		if !canAccessMaintenanceVehicle(userID, input.VehicleID) { c.JSON(http.StatusForbidden, gin.H{"error": "Fahrzeug nicht freigegeben"}); return }
		task := db.MaintenanceTask{OwnerUserID:userID, VehicleID:input.VehicleID, Title:strings.TrimSpace(input.Title), Notes:input.Notes, DueDate:input.DueDate, DueOdo:input.DueOdo}
		if err := db.DB.Create(&task).Error; err != nil { c.JSON(500, gin.H{"error": "Wartung konnte nicht gespeichert werden"}); return }
		c.JSON(201, task)
	})
	router.PUT("/maintenance/:id", func(c *gin.Context) {
		var input maintenanceInput
		if c.ShouldBindJSON(&input) != nil || strings.TrimSpace(input.Title) == "" || len(input.Title) > 120 || len(input.Notes) > 2000 || (input.DueOdo == nil && input.DueDate == nil) || (input.DueOdo != nil && *input.DueOdo < 0) {
			c.JSON(422, gin.H{"error": "Ungültige Wartungsdaten"}); return
		}
		userID := c.MustGet("userId").(string)
		if !canAccessMaintenanceVehicle(userID, input.VehicleID) { c.JSON(403, gin.H{"error": "Fahrzeug nicht freigegeben"}); return }
		var task db.MaintenanceTask
		if db.DB.Where("id = ? AND owner_user_id = ?", c.Param("id"), userID).First(&task).Error != nil { c.Status(404); return }
		task.VehicleID = input.VehicleID
		task.Title = strings.TrimSpace(input.Title)
		task.Notes = input.Notes
		task.DueDate = input.DueDate
		task.DueOdo = input.DueOdo
		if input.Completed && task.CompletedAt == nil { now := time.Now(); task.CompletedAt = &now }
		if !input.Completed { task.CompletedAt = nil }
		if err := db.DB.Save(&task).Error; err != nil { c.JSON(500, gin.H{"error":"Speichern fehlgeschlagen"}); return }
		c.JSON(200, task)
	})
	router.DELETE("/maintenance/:id", func(c *gin.Context) {
		result := db.DB.Where("id = ? AND owner_user_id = ?", c.Param("id"), c.MustGet("userId").(string)).Delete(&db.MaintenanceTask{})
		if result.Error != nil { c.Status(500); return }
		if result.RowsAffected == 0 { c.Status(404); return }
		c.Status(204)
	})
}
