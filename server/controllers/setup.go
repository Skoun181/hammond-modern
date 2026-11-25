package controllers

import (
	"net/http"

	"github.com/akhilrex/hammond/db"
	"github.com/gin-gonic/gin"
)

func RegisterSetupController(router *gin.RouterGroup) {
	router.GET("/system/status", appInitialized)
	// Preserve an actionable response for older clients accessing a retired route.
	retired := func(c *gin.Context) {
		c.JSON(http.StatusGone, gin.H{"error": "Use the local --import-clarkson command"})
	}
	router.POST("/clarkson/check", retired)
	router.POST("/clarkson/migrate", retired)
}

func appInitialized(c *gin.Context) {
	canInitialize, err := db.CanInitializeSystem()
	if err != nil {
		c.Status(http.StatusServiceUnavailable)
		return
	}
	c.JSON(http.StatusOK, gin.H{"initialized": !canInitialize})
}
