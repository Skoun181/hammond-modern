package controllers

import (
	"net/http"

	"github.com/akhilrex/hammond/db"
	"github.com/gin-gonic/gin"
)

func vehicleAccess(ownerOnly bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := db.DB.Model(&db.UserVehicle{}).
			Where("vehicle_id = ? AND user_id = ?", c.Param("id"), c.GetString("userId"))
		if ownerOnly {
			query = query.Where("is_owner = ?", true)
		}
		var count int64
		if err := query.Count(&count).Error; err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		if count == 0 {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		c.Next()
	}
}

func belongsToVehicle(c *gin.Context, model interface{}) bool {
	var count int64
	query := db.DB.Model(model).Where("id = ? AND vehicle_id = ?", c.Param("subId"), c.Param("id"))
	if err := query.Count(&count).Error; err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return false
	}
	if count == 0 {
		c.AbortWithStatus(http.StatusNotFound)
		return false
	}
	return true
}
