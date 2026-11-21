package controllers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/akhilrex/hammond/db"
	"github.com/akhilrex/hammond/service"
	"github.com/gin-gonic/gin"
)

// Blocks all other API requests during a restore. The restore handler obtains
// the exclusive lock after authentication and upload validation.
func PortableDataGuard() gin.HandlerFunc {
	return func(c *gin.Context) {
		checkJournal := func() bool {
			_, err := os.Stat(filepath.Join(os.Getenv("CONFIG"), "restore-safety", "in-progress.json"))
			if err == nil || !os.IsNotExist(err) {
				c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "Unterbrochene Wiederherstellung erkannt. Bitte Container für automatische Recovery neu starten."})
				return false
			}
			return true
		}
		if c.Request.URL.Path == "/api/backup/restore" {
			if checkJournal() {
				c.Next()
			}
			return
		}
		service.PortableDataLock.RLock()
		defer service.PortableDataLock.RUnlock()
		// Check only after obtaining the read lock; a concurrent restore may
		// have just left an incomplete journal while we were waiting.
		if !checkJournal() {
			return
		}
		c.Next()
	}
}

func RegisterPortableBackupController(router *gin.RouterGroup) {
	router.GET("/backup/export", ShouldBeAdmin(), portableExport)
	router.POST("/backup/preview", ShouldBeAdmin(), portablePreview)
	router.POST("/backup/restore", ShouldBeAdmin(), portableRestore)
	router.POST("/backup/users/:id/password", ShouldBeAdmin(), portableResetPassword)
}

func portableError(c *gin.Context, status int, err error) {
	log.Printf("portable backup operation failed: %v", err)
	c.JSON(status, gin.H{"error": err.Error()})
}

func portableExport(c *gin.Context) {
	file, _, err := service.CreatePortableZIP()
	if err != nil {
		portableError(c, http.StatusInternalServerError, err)
		return
	}
	defer os.Remove(file)
	filename := "hammond-modern-backup-" + time.Now().Format("2006-01-02-150405") + ".zip"
	c.Header("Cache-Control", "no-store")
	c.Header("X-Content-Type-Options", "nosniff")
	c.FileAttachment(file, filename)
}

func portableUpload(c *gin.Context) (string, func(), error) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, service.PortableMaxZIP+(2<<20))
	form, err := c.FormFile("file")
	if err != nil {
		return "", nil, fmt.Errorf("Bitte eine gültige ZIP-Datei auswählen (maximal 512 MiB)")
	}
	if form.Size <= 0 || form.Size > service.PortableMaxZIP || !strings.HasSuffix(strings.ToLower(form.Filename), ".zip") {
		return "", nil, fmt.Errorf("Erlaubt sind nur ZIP-Dateien bis 512 MiB")
	}
	input, err := form.Open()
	if err != nil {
		return "", nil, err
	}
	defer input.Close()
	temp, err := os.CreateTemp("", "hm-backup-upload-*.zip")
	if err != nil {
		return "", nil, err
	}
	cleanup := func() { os.Remove(temp.Name()) }
	if err = temp.Chmod(0600); err != nil {
		temp.Close()
		cleanup()
		return "", nil, err
	}
	n, err := io.Copy(temp, io.LimitReader(input, service.PortableMaxZIP+1))
	if closeErr := temp.Close(); err == nil {
		err = closeErr
	}
	if err != nil || n > service.PortableMaxZIP {
		cleanup()
		return "", nil, fmt.Errorf("Upload fehlgeschlagen oder Datei zu groß")
	}
	return temp.Name(), cleanup, nil
}

func portablePreview(c *gin.Context) {
	file, cleanup, err := portableUpload(c)
	if err != nil {
		portableError(c, http.StatusBadRequest, err)
		return
	}
	defer cleanup()
	user := c.MustGet("userModel").(db.User)
	summary, err := service.InspectPortableZIP(file, user)
	if err != nil {
		portableError(c, http.StatusUnprocessableEntity, err)
		return
	}
	c.JSON(http.StatusOK, summary)
}

func portableRestore(c *gin.Context) {
	if c.PostForm("confirm") != "ERSETZEN" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Zur Bestätigung ERSETZEN eingeben."})
		return
	}
	file, cleanup, err := portableUpload(c)
	if err != nil {
		portableError(c, http.StatusBadRequest, err)
		return
	}
	defer cleanup()
	user := c.MustGet("userModel").(db.User)
	service.PortableDataLock.Lock()
	defer service.PortableDataLock.Unlock()
	// Recheck identity after acquiring the lock: another concurrent restore
	// could have replaced the user database while this request was waiting.
	var current db.User
	if lookupErr := db.DB.Where("id=?", user.ID).First(&current).Error; lookupErr != nil || current.Role != db.ADMIN || current.IsDisabled {
		c.JSON(http.StatusForbidden, gin.H{"error": "Administratorberechtigung ist nicht mehr gültig. Bitte erneut anmelden."})
		return
	}
	summary, err := service.RestorePortableZIP(file, current)
	if err != nil {
		portableError(c, http.StatusUnprocessableEntity, err)
		return
	}
	c.JSON(http.StatusOK, summary)
}

// The only password carried into a restored backup is the destination admin's
// existing bcrypt hash. Other users can be re-enabled with a new password here.
func portableResetPassword(c *gin.Context) {
	var req struct {
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || len(req.Password) < 12 || len(req.Password) > 72 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bitte ein Passwort mit 12 bis 72 Zeichen angeben."})
		return
	}
	var user db.User
	if err := db.DB.Where("id=?", c.Param("id")).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Benutzer nicht gefunden."})
		return
	}
	if err := user.SetPassword(req.Password); err != nil {
		portableError(c, http.StatusBadRequest, err)
		return
	}
	if err := db.DB.Model(&user).Update("password", user.Password).Error; err != nil {
		portableError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}
