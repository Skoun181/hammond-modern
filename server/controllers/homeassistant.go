package controllers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/akhilrex/hammond/db"
	"github.com/akhilrex/hammond/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const homeAssistantTokenPrefix = "hmha_"
const homeAssistantTokenLength = 48 // prefix plus 43 base64url chars (32 random bytes)

// Token hashes are used only for lookup, never returned or logged.
func homeAssistantHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func newHomeAssistantToken() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return homeAssistantTokenPrefix + base64.RawURLEncoding.EncodeToString(raw), nil
}

func homeAssistantUserFromHeader(value string) (*db.User, error) {
	fields := strings.Fields(value)
	if len(fields) != 2 || !strings.EqualFold(fields[0], "Bearer") ||
		!strings.HasPrefix(fields[1], homeAssistantTokenPrefix) || len(fields[1]) != homeAssistantTokenLength {
		return nil, gorm.ErrRecordNotFound
	}
	// The actual token material is never stored in the database.
	var key db.HomeAssistantKey
	if err := db.DB.Where("token_hash = ?", homeAssistantHash(fields[1])).First(&key).Error; err != nil {
		return nil, err
	}
	if key.OwnerUserID == "" {
		return nil, gorm.ErrRecordNotFound
	}
	var user db.User
	if err := db.DB.Where("id = ? AND deleted_at IS NULL", key.OwnerUserID).First(&user).Error; err != nil {
		return nil, err
	}
	if user.ID == "" || user.IsDisabled {
		return nil, gorm.ErrRecordNotFound
	}
	return &user, nil
}

func homeAssistantNoCache(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	c.Header("Pragma", "no-cache")
	c.Header("X-Content-Type-Options", "nosniff")
}

// The following endpoints require a regular logged-in Hammond user (JWT).
// A new token replaces the previous one, which is invalid immediately.
func RegisterHomeAssistantKeyController(router *gin.RouterGroup) {
	group := router.Group("/integrations/homeassistant/key")
	group.Use(planningAuth)
	group.GET("", func(c *gin.Context) {
		homeAssistantNoCache(c)
		userID := c.GetString("userId")
		var key db.HomeAssistantKey
		err := db.DB.Where("owner_user_id = ?", userID).First(&key).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusOK, gin.H{"configured": false})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Status konnte nicht geladen werden"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"configured": key.TokenHash != "", "createdAt": key.CreatedAt})
	})
	group.POST("", func(c *gin.Context) {
		homeAssistantNoCache(c)
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1024)
		token, err := newHomeAssistantToken()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Schlüssel konnte nicht erstellt werden"})
			return
		}
		userID := c.GetString("userId")
		key := db.HomeAssistantKey{OwnerUserID: userID, TokenHash: homeAssistantHash(token), CreatedAt: time.Now().UTC()}
		err = db.DB.Transaction(func(tx *gorm.DB) error {
			if e := tx.Where("owner_user_id = ?", userID).Delete(&db.HomeAssistantKey{}).Error; e != nil {
				return e
			}
			return tx.Create(&key).Error
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Schlüssel konnte nicht gespeichert werden"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"token": token, "createdAt": key.CreatedAt})
	})
	group.DELETE("", func(c *gin.Context) {
		homeAssistantNoCache(c)
		if err := db.DB.Where("owner_user_id = ?", c.GetString("userId")).Delete(&db.HomeAssistantKey{}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Schlüssel konnte nicht widerrufen werden"})
			return
		}
		c.Status(http.StatusNoContent)
	})
}

type homeAssistantTask struct {
	ID string `json:"id"`
	VehicleID string `json:"vehicle_id"`
	VehicleName string `json:"vehicle_name"`
	Title string `json:"title"`
	Severity string `json:"severity"`
	DueDate *string `json:"due_date"`
	DueOdo *int `json:"due_odometer"`
	CurrentOdo *int `json:"current_odometer"`
	DaysLeft *int `json:"days_left"`
	KmLeft *int `json:"km_left"`
}

type homeAssistantSummary struct {
	Total int `json:"total"`
	Overdue int `json:"overdue"`
	Due int `json:"due"`
	Soon int `json:"soon"`
	Planned int `json:"planned"`
}

func homeAssistantMaintenance(userID string, now time.Time) (homeAssistantSummary, []homeAssistantTask, error) {
	var summary homeAssistantSummary
	var tasks []db.MaintenanceTask
	err := db.DB.Where("owner_user_id = ? AND completed_at IS NULL AND deleted_at IS NULL", userID).
		Order("created_at ASC").Limit(500).Find(&tasks).Error
	if err != nil {
		return summary, nil, err
	}
	type vehicleData struct {
		name string
		odo *int
		allowed bool
	}
	vehicles := map[string]vehicleData{}
	out := make([]homeAssistantTask, 0, len(tasks))
	for _, task := range tasks {
		vehicle, ok := vehicles[task.VehicleID]
		if !ok {
			allowed, accessErr := service.PlanningAccess(db.DB, userID, task.VehicleID)
			if accessErr != nil {
				return summary, nil, accessErr
			}
			vehicle = vehicleData{allowed: allowed}
			if allowed {
				var model db.Vehicle
				if e := db.DB.Where("id = ? AND deleted_at IS NULL", task.VehicleID).First(&model).Error; e != nil {
					return summary, nil, e
				}
				vehicle.name = strings.TrimSpace(model.Nickname)
				if vehicle.name == "" {
					vehicle.name = strings.TrimSpace(model.Make + " " + model.Model)
				}
				if vehicle.name == "" {
					vehicle.name = strings.TrimSpace(model.Registration)
				}
				if vehicle.name == "" {
					vehicle.name = "Fahrzeug"
				}
				vehicle.odo, err = service.PlanningOdo(db.DB, task.VehicleID, now)
				if err != nil {
					return summary, nil, err
				}
			}
			vehicles[task.VehicleID] = vehicle
		}
		if !vehicle.allowed {
			continue
		}
		rule, e := getMaintenanceRule(db.DB, task.ID)
		if e != nil {
			return summary, nil, e
		}
		status := maintenanceStatus(task, rule, vehicle.odo, now)
		item := homeAssistantTask{
			ID: task.ID, VehicleID: task.VehicleID, VehicleName: vehicle.name,
			Title: task.Title, Severity: status.Severity, DueOdo: task.DueOdo,
			CurrentOdo: vehicle.odo, DaysLeft: status.DaysLeft, KmLeft: status.KmLeft,
		}
		if task.DueDate != nil {
			date := task.DueDate.Format("2006-01-02")
			item.DueDate = &date
		}
		out = append(out, item)
		switch status.Severity {
		case "overdue":
			summary.Overdue++
		case "due":
			summary.Due++
		case "soon":
			summary.Soon++
		default:
			summary.Planned++
		}
		summary.Total++
	}
	priority := map[string]int{"overdue": 0, "due": 1, "soon": 2, "planned": 3}
	sort.SliceStable(out, func(i, j int) bool {
		if priority[out[i].Severity] != priority[out[j].Severity] {
			return priority[out[i].Severity] < priority[out[j].Severity]
		}
		if out[i].DueDate != nil && out[j].DueDate != nil && *out[i].DueDate != *out[j].DueDate {
			return *out[i].DueDate < *out[j].DueDate
		}
		return out[i].VehicleName + out[i].Title < out[j].VehicleName + out[j].Title
	})
	return summary, out, nil
}

// Registers without normal JWT middleware. Auth is ONLY a read-only API key.
func RegisterHomeAssistantReadController(router *gin.RouterGroup) {
	router.GET("/integrations/homeassistant/maintenance", func(c *gin.Context) {
		homeAssistantNoCache(c)
		user, err := homeAssistantUserFromHeader(c.GetHeader("Authorization"))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Ungültiger oder widerrufener Home-Assistant-Schlüssel"})
			return
		}
		summary, tasks, err := homeAssistantMaintenance(user.ID, time.Now())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Wartungen konnten nicht gelesen werden"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"version": 1, "account_id": user.ID, "summary": summary, "tasks": tasks,
		})
	})
}
