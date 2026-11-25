package controllers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/akhilrex/hammond/common"
	"github.com/akhilrex/hammond/db"
	"github.com/akhilrex/hammond/models"
	"github.com/akhilrex/hammond/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterAnonController(router *gin.RouterGroup) {
	router.POST("/login", userLogin)
	router.POST("/auth/initialize", initializeSystem)
	router.POST("/refresh", refresh)
	router.POST("/logout", logout)
}

func RegisterAuthController(router *gin.RouterGroup) {
	router.GET("/me", me)
	router.POST("/register", ShouldBeAdmin(), userRegister)
	router.POST("/changePassword", changePassword)
}

func ShouldBeAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		model, _ := c.Get("userModel")
		user, _ := model.(db.User)
		if user.ID == "" || user.Role != db.ADMIN || user.IsDisabled {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Administratorrechte erforderlich"})
			return
		}
		c.Next()
	}
}

func me(c *gin.Context) {
	user, err := service.GetUserById(c.GetString("userId"))
	if err != nil {
		c.Status(http.StatusUnauthorized)
		return
	}
	c.JSON(http.StatusOK, user)
}

func userRegister(c *gin.Context) {
	var request models.RegisterRequest
	if err := c.ShouldBind(&request); err != nil {
		c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
		return
	}
	if request.Role == nil || (*request.Role != db.ADMIN && *request.Role != db.USER) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Ungültige Benutzerrolle"})
		return
	}
	if err := service.CreateUser(&request, *request.Role); err != nil {
		c.JSON(http.StatusUnprocessableEntity, common.NewError("database", err))
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true})
}

var errInitialized = errors.New("system already initialized")

func initializeSystem(c *gin.Context) {
	if !sessionOrigin(c) {
		c.Status(http.StatusForbidden)
		return
	}
	available, err := db.CanInitializeSystem()
	if err != nil {
		c.Status(http.StatusServiceUnavailable)
		return
	}
	if !available {
		c.Status(http.StatusConflict)
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 8192)
	var request models.RegisterRequest
	if err := c.ShouldBind(&request); err != nil {
		c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
		return
	}
	if request.DistanceUnit == nil || (*request.DistanceUnit != db.KILOMETERS && *request.DistanceUnit != db.MILES) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Ungültige Distanzeinheit"})
		return
	}
	user := db.User{Email: strings.ToLower(request.Email), Name: request.Name, Role: db.ADMIN,
		Currency: request.Currency, DistanceUnit: *request.DistanceUnit, DateFormat: "MM/dd/yyyy"}
	if err := user.SetPassword(request.Password); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	err = db.DB.Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Unscoped().Model(&db.User{}).Count(&count).Error; err != nil {
			return err
		}
		if count != 0 {
			return errInitialized
		}
		if err := tx.Create(&user).Error; err != nil {
			return err
		}
		var settings db.Setting
		if err := tx.First(&settings).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		settings.Currency, settings.DistanceUnit = request.Currency, *request.DistanceUnit
		return tx.Save(&settings).Error
	})
	if err != nil {
		if errors.Is(err, errInitialized) {
			c.Status(http.StatusConflict)
		} else {
			c.Status(http.StatusServiceUnavailable)
		}
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true})
}

func sendLogin(c *gin.Context, user *db.User, tokens service.SessionTokens) {
	sessionCookies(c, tokens, true)
	c.Header("Cache-Control", "no-store")
	response := models.LoginResponse{Name: user.Name, Email: user.Email, Role: user.RoleDetail().Long}
	if c.GetHeader("Origin") == "" && c.GetHeader("Sec-Fetch-Site") == "" {
		response.Token, response.RefreshToken = tokens.Access, tokens.Refresh
	}
	c.JSON(http.StatusOK, response)
}

func userLogin(c *gin.Context) {
	if !sessionOrigin(c) {
		c.Status(http.StatusForbidden)
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 8192)
	var loginRequest models.LoginRequest
	if err := c.ShouldBind(&loginRequest); err != nil {
		c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
		return
	}
	user, err := db.FindOneUser(&db.User{Email: strings.ToLower(loginRequest.Email)})
	if err != nil || user.IsDisabled || user.CheckPassword(loginRequest.Password) != nil {
		c.JSON(http.StatusForbidden, common.NewError("login", errors.New("invalid login")))
		return
	}
	tokens, err := service.NewSession(&user)
	if err != nil {
		c.Status(http.StatusServiceUnavailable)
		return
	}
	sendLogin(c, &user, tokens)
}

func refresh(c *gin.Context) {
	if !sessionOrigin(c) {
		c.Status(http.StatusForbidden)
		return
	}
	token, _ := c.Cookie(refreshCookie)
	if token == "" {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 8192)
		var request struct {
			Token string `json:"refreshToken"`
		}
		if err := c.ShouldBindJSON(&request); err != nil {
			c.Status(http.StatusUnauthorized)
			return
		}
		token = request.Token
	}
	user, tokens, err := service.RotateSession(token)
	if err != nil {
		c.Status(http.StatusUnauthorized)
		return
	}
	sendLogin(c, user, tokens)
}

func logout(c *gin.Context) {
	if !sessionOrigin(c) {
		c.Status(http.StatusForbidden)
		return
	}
	access := bearer(c)
	if access == "" {
		access, _ = c.Cookie(accessCookie)
	}
	refreshToken, _ := c.Cookie(refreshCookie)
	if err := service.RevokeSession(access, refreshToken); err != nil {
		c.Status(http.StatusServiceUnavailable)
		return
	}
	sessionCookies(c, service.SessionTokens{}, false)
	c.Header("Cache-Control", "no-store")
	c.Status(http.StatusNoContent)
}

func changePassword(c *gin.Context) {
	var request models.ChangePasswordRequest
	if err := c.ShouldBind(&request); err != nil {
		c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
		return
	}
	user, err := service.GetUserById(c.GetString("userId"))
	if err != nil || user.CheckPassword(request.OldPassword) != nil {
		c.JSON(http.StatusForbidden, common.NewError("changePassword", errors.New("incorrect old password")))
		return
	}
	success, err := service.UpdatePassword(user.ID, request.NewPassword)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	sessionCookies(c, service.SessionTokens{}, false)
	c.JSON(http.StatusOK, success)
}
