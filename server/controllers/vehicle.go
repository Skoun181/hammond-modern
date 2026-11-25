package controllers

import (
	"errors"
	"net/http"
	"os"

	"github.com/akhilrex/hammond/common"
	"github.com/akhilrex/hammond/db"
	"github.com/akhilrex/hammond/models"
	"github.com/akhilrex/hammond/service"
	"github.com/gin-gonic/gin"
)

func RegisterVehicleController(router *gin.RouterGroup) {
	router.POST("/vehicles", createVehicle)
	router.GET("/vehicles", getAllVehicles)
	router.GET("/vehicles/:id", vehicleAccess(false), getVehicleById)
	router.PUT("/vehicles/:id", vehicleAccess(true), updateVehicle)
	router.DELETE("/vehicles/:id", vehicleAccess(true), deleteVehicle)
	router.GET("/vehicles/:id/stats", vehicleAccess(false), getVehicleStats)
	router.GET("/vehicles/:id/users", vehicleAccess(false), getVehicleUsers)
	router.POST("/vehicles/:id/users/:subId", vehicleAccess(true), shareVehicle)
	router.DELETE("/vehicles/:id/users/:subId", vehicleAccess(true), unshareVehicle)
	router.POST("/vehicles/:id/users/:subId/transfer", vehicleAccess(true), transferVehicle)

	router.GET("/me/vehicles", getMyVehicles)
	router.GET("/me/stats", getMystats)

	router.GET("/vehicles/:id/fillups", vehicleAccess(false), getFillupsByVehicleId)
	router.GET("/vehicles/:id/fuelSubTypes", vehicleAccess(false), getFuelSubTypesByVehicleId)
	router.POST("/vehicles/:id/fillups", vehicleAccess(false), createFillup)
	router.GET("/vehicles/:id/fillups/:subId", vehicleAccess(false), getFillupById)
	router.PUT("/vehicles/:id/fillups/:subId", vehicleAccess(false), updateFillup)
	router.DELETE("/vehicles/:id/fillups/:subId", vehicleAccess(false), deleteFillup)

	router.GET("/vehicles/:id/expenses", vehicleAccess(false), getExpensesByVehicleId)
	router.POST("/vehicles/:id/expenses", vehicleAccess(false), createExpense)
	router.GET("/vehicles/:id/expenses/:subId", vehicleAccess(false), getExpenseById)
	router.PUT("/vehicles/:id/expenses/:subId", vehicleAccess(false), updateExpense)
	router.DELETE("/vehicles/:id/expenses/:subId", vehicleAccess(false), deleteExpense)

	router.POST("/vehicles/:id/attachments", vehicleAccess(false), createVehicleAttachment)
	router.GET("/vehicles/:id/attachments", vehicleAccess(false), getVehicleAttachments)
}

func createVehicle(c *gin.Context) {
	var request models.CreateVehicleRequest
	if err := c.ShouldBind(&request); err != nil {
		c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
		return
	}
	vehicle, err := service.CreateVehicle(request, c.MustGet("userId").(string))
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, common.NewError("createVehicle", err))
		return
	}
	c.JSON(http.StatusCreated, vehicle)
}
func getVehicleById(c *gin.Context) {
	var searchByIdQuery models.SearchByIdQuery

	if c.ShouldBindUri(&searchByIdQuery) == nil {
		vehicle, err := service.GetVehicleById(searchByIdQuery.Id)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("getVehicleById", err))
			return
		}
		c.JSON(http.StatusOK, vehicle)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
	}
}
func updateVehicle(c *gin.Context) {
	var searchByIdQuery models.SearchByIdQuery
	var updateVehicleModel models.UpdateVehicleRequest
	if err := c.ShouldBindUri(&searchByIdQuery); err == nil {
		if err := c.ShouldBind(&updateVehicleModel); err == nil {
			err := service.UpdateVehicle(searchByIdQuery.Id, updateVehicleModel)
			if err != nil {
				c.JSON(http.StatusUnprocessableEntity, common.NewError("getVehicleById", err))
				return
			}
			c.JSON(http.StatusOK, gin.H{})
		} else {
			c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
		}
	} else {
		c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
	}
}
func getAllVehicles(c *gin.Context) {
	vehicles, err := service.GetUserVehicles(c.GetString("userId"))
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, common.NewError("getVehicleById", err))
		return
	}
	c.JSON(200, vehicles)

}
func getMyVehicles(c *gin.Context) {
	vehicles, err := service.GetUserVehicles(c.MustGet("userId").(string))
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, common.NewError("getMyVehicles", err))
		return
	}
	c.JSON(200, vehicles)

}

func getFillupsByVehicleId(c *gin.Context) {

	var searchByIdQuery models.SearchByIdQuery

	if err := c.ShouldBindUri(&searchByIdQuery); err == nil {

		fillups, err := service.GetFillupsByVehicleId(searchByIdQuery.Id)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("getFillupsByVehicleId", err))
			return
		}
		c.JSON(http.StatusOK, fillups)
	} else {
		c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
	}
}
func getFuelSubTypesByVehicleId(c *gin.Context) {

	var searchByIdQuery models.SearchByIdQuery

	if err := c.ShouldBindUri(&searchByIdQuery); err == nil {

		fuelSubtypes, err := service.GetDistinctFuelSubtypesForVehicle(searchByIdQuery.Id)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("getFuelSubTypesByVehicleId", err))
			return
		}
		c.JSON(http.StatusOK, fuelSubtypes)
	} else {
		c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
	}
}

func getExpensesByVehicleId(c *gin.Context) {

	var searchByIdQuery models.SearchByIdQuery

	if err := c.ShouldBindUri(&searchByIdQuery); err == nil {

		data, err := service.GetExpensesByVehicleId(searchByIdQuery.Id)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("getExpensesByVehicleId", err))
			return
		}
		c.JSON(http.StatusOK, data)
	} else {
		c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
	}
}

func createFillup(c *gin.Context) {
	var request models.CreateFillupRequest
	var searchByIdQuery models.SearchByIdQuery

	if err := c.ShouldBindUri(&searchByIdQuery); err == nil {
		if err := c.ShouldBind(&request); err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
			return
		}
		if request.DriverID != "" {
			var count int64
			if err := db.DB.Model(&db.Driver{}).Where("id = ? AND owner_user_id = ?", request.DriverID, c.MustGet("userId").(string)).Count(&count).Error; err != nil || count == 0 {
				c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Fahrer nicht gefunden"})
				return
			}
		}
		request.VehicleID, request.UserID = searchByIdQuery.Id, c.GetString("userId")
		fillup, err := service.CreateFillup(request)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("createFillup", err))
			return
		}
		c.JSON(http.StatusCreated, fillup)
	} else {
		c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
	}
}

func createExpense(c *gin.Context) {
	var request models.CreateExpenseRequest
	var searchByIdQuery models.SearchByIdQuery

	if err := c.ShouldBindUri(&searchByIdQuery); err == nil {
		if err := c.ShouldBind(&request); err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
			return
		}
		if !validInvoiceForUser(request.InvoiceAttachmentID, c.MustGet("userId").(string)) {
			c.JSON(422, gin.H{"error": "Rechnung gehört nicht zum angemeldeten Benutzer"})
			return
		}
		request.VehicleID, request.UserID = searchByIdQuery.Id, c.GetString("userId")
		expense, err := service.CreateExpense(request)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("createExpense", err))
			return
		}
		c.JSON(http.StatusCreated, expense)
	} else {
		c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
	}
}

func updateExpense(c *gin.Context) {
	if !belongsToVehicle(c, &db.Expense{}) {
		return
	}
	var searchByIdQuery models.SubItemQuery
	var updateExpenseModel models.UpdateExpenseRequest
	if err := c.ShouldBindUri(&searchByIdQuery); err == nil {
		if err := c.ShouldBind(&updateExpenseModel); err == nil {
			if !validInvoiceForUser(updateExpenseModel.InvoiceAttachmentID, c.MustGet("userId").(string)) {
				c.JSON(422, gin.H{"error": "Rechnung gehört nicht zum angemeldeten Benutzer"})
				return
			}
			updateExpenseModel.VehicleID = searchByIdQuery.Id
			err := service.UpdateExpense(searchByIdQuery.SubId, updateExpenseModel)
			if err != nil {
				c.JSON(http.StatusUnprocessableEntity, common.NewError("getExpenseById", err))
				return
			}
			c.JSON(http.StatusOK, gin.H{})
		} else {
			c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
		}
	} else {
		c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
	}
}
func updateFillup(c *gin.Context) {
	if !belongsToVehicle(c, &db.Fillup{}) {
		return
	}
	var searchByIdQuery models.SubItemQuery
	var updateFillupModel models.UpdateFillupRequest
	if err := c.ShouldBindUri(&searchByIdQuery); err == nil {
		if err := c.ShouldBind(&updateFillupModel); err == nil {
			updateFillupModel.VehicleID = searchByIdQuery.Id
			err := service.UpdateFillup(searchByIdQuery.SubId, updateFillupModel)
			if err != nil {
				c.JSON(http.StatusUnprocessableEntity, common.NewError("getFillupById", err))
				return
			}
			c.JSON(http.StatusOK, gin.H{})
		} else {
			c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
		}
	} else {
		c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
	}
}

func deleteExpense(c *gin.Context) {
	if !belongsToVehicle(c, &db.Expense{}) {
		return
	}
	var searchByIdQuery models.SubItemQuery

	if err := c.ShouldBindUri(&searchByIdQuery); err == nil {

		err := service.DeleteExpenseById(searchByIdQuery.SubId)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("getExpenseById", err))
			return
		}
		c.JSON(http.StatusOK, gin.H{})

	} else {
		c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
	}
}
func deleteFillup(c *gin.Context) {
	if !belongsToVehicle(c, &db.Fillup{}) {
		return
	}
	var searchByIdQuery models.SubItemQuery

	if err := c.ShouldBindUri(&searchByIdQuery); err == nil {

		err := service.DeleteFillupById(searchByIdQuery.SubId)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("getFillupById", err))
			return
		}
		c.JSON(http.StatusOK, gin.H{})

	} else {
		c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
	}
}

func getExpenseById(c *gin.Context) {
	if !belongsToVehicle(c, &db.Expense{}) {
		return
	}
	var searchByIdQuery models.SubItemQuery

	if err := c.ShouldBindUri(&searchByIdQuery); err == nil {

		obj, err := service.GetExpenseById(searchByIdQuery.SubId)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("getExpenseById", err))
			return
		}
		c.JSON(http.StatusOK, obj)

	} else {
		c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
	}
}
func getFillupById(c *gin.Context) {
	if !belongsToVehicle(c, &db.Fillup{}) {
		return
	}
	var searchByIdQuery models.SubItemQuery

	if err := c.ShouldBindUri(&searchByIdQuery); err == nil {

		obj, err := service.GetFillupById(searchByIdQuery.SubId)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("getFillupById", err))
			return
		}
		c.JSON(http.StatusOK, obj)

	} else {
		c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
	}
}

func createVehicleAttachment(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxAttachment+1024*1024)
	var searchByIdQuery models.SearchByIdQuery
	var dataModel models.CreateVehicleAttachmentModel
	if err := c.ShouldBindUri(&searchByIdQuery); err == nil {
		if err := c.ShouldBind(&dataModel); err == nil {
			vehicle, err := service.GetVehicleById(searchByIdQuery.Id)
			if err != nil {
				c.JSON(http.StatusUnprocessableEntity, common.NewError("createVehicleAttachment", err))
				return
			}
			file, err := c.FormFile("file")
			if err != nil {
				c.Status(http.StatusBadRequest)
				return
			}
			kind, err := attachmentType(file)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			attachment, err := storeAttachment(c, file, kind)
			if err != nil {
				c.JSON(http.StatusUnprocessableEntity, common.NewError("createVehicleAttachment", err))
				return
			}
			err = service.CreateVehicleAttachment(vehicle.ID, attachment.ID, dataModel.Title)
			if err != nil {
				if db.DB.Unscoped().Delete(attachment).Error == nil {
					_ = os.Remove(attachment.Path)
				}
				c.JSON(http.StatusUnprocessableEntity, common.NewError("createVehicleAttachment", err))
				return
			}
			c.JSON(http.StatusOK, gin.H{})
		} else {
			c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
		}
	} else {
		c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
	}
}

func getVehicleAttachments(c *gin.Context) {
	var searchByIdQuery models.SearchByIdQuery

	if err := c.ShouldBindUri(&searchByIdQuery); err == nil {

		vehicle, err := service.GetVehicleById(searchByIdQuery.Id)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("createVehicleAttachment", err))
			return
		}

		attachments, err := service.GetVehicleAttachments(vehicle.ID)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("createVehicleAttachment", err))
			return
		}
		c.JSON(http.StatusOK, attachments)

	} else {
		c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
	}
}

func getVehicleUsers(c *gin.Context) {
	var searchByIdQuery models.SearchByIdQuery

	if err := c.ShouldBindUri(&searchByIdQuery); err == nil {

		vehicle, err := service.GetVehicleById(searchByIdQuery.Id)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("getVehicleUsers", err))
			return
		}
		data, err := service.GetVehicleUsers(vehicle.ID)

		var model []models.UserVehicleSimpleModel

		for _, item := range *data {
			model = append(model, models.UserVehicleSimpleModel{
				ID:        item.ID,
				UserID:    item.UserID,
				VehicleID: item.VehicleID,
				IsOwner:   item.IsOwner,
				Name:      item.User.Name,
			})
		}

		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("getVehicleUsers", err))
			return
		}

		c.JSON(http.StatusOK, model)

	} else {
		c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
	}
}
func deleteVehicle(c *gin.Context) {
	var searchByIdQuery models.SearchByIdQuery

	if err := c.ShouldBindUri(&searchByIdQuery); err == nil {

		canDelete, err := service.CanDeleteVehicle(searchByIdQuery.Id, c.MustGet("userId").(string))
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("shareVehicle", err))
			return
		}
		if !canDelete {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("shareVehicle", errors.New("you are not allowed to delete this vehicle")))
			return
		}
		err = service.DeleteVehicle(searchByIdQuery.Id)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("shareVehicle", err))
			return
		}
		c.JSON(http.StatusOK, gin.H{})

	} else {
		c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
	}
}
func shareVehicle(c *gin.Context) {
	var searchByIdQuery models.SubItemQuery

	if err := c.ShouldBindUri(&searchByIdQuery); err == nil {

		err := service.ShareVehicle(searchByIdQuery.Id, searchByIdQuery.SubId)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("shareVehicle", err))
			return
		}
		c.JSON(http.StatusOK, gin.H{})

	} else {
		c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
	}
}
func transferVehicle(c *gin.Context) {
	var searchByIdQuery models.SubItemQuery

	if err := c.ShouldBindUri(&searchByIdQuery); err == nil {

		err := service.TransferVehicle(searchByIdQuery.Id, c.MustGet("userId").(string), searchByIdQuery.SubId)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("shareVehicle", err))
			return
		}
		c.JSON(http.StatusOK, gin.H{})

	} else {
		c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
	}
}
func unshareVehicle(c *gin.Context) {
	var searchByIdQuery models.SubItemQuery

	if err := c.ShouldBindUri(&searchByIdQuery); err == nil {

		err := service.UnshareVehicle(searchByIdQuery.Id, searchByIdQuery.SubId)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("shareVehicle", err))
			return
		}
		c.JSON(http.StatusOK, gin.H{})

	} else {
		c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
	}
}

func getVehicleStats(c *gin.Context) {
	var searchByIdQuery models.SearchByIdQuery

	if err := c.ShouldBindUri(&searchByIdQuery); err == nil {

		vehicle, err := service.GetVehicleById(searchByIdQuery.Id)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("getVehicleState", err))
			return
		}

		model := models.VehicleStatsModel{}

		c.JSON(http.StatusOK, model.SetStats(&vehicle.Fillups, &vehicle.Expenses))

	} else {
		c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
	}
}

func getMystats(c *gin.Context) {
	var model models.UserStatsQueryModel
	if err := c.ShouldBind(&model); err == nil {

		stats, err := service.GetUserStats(c.MustGet("userId").(string), model)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("getMyVehicles", err))
			return
		}
		c.JSON(200, stats)
	} else {
		c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
	}

}
