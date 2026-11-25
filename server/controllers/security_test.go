package controllers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/akhilrex/hammond/db"
	"github.com/akhilrex/hammond/service"
	"github.com/gin-gonic/gin"
)

func TestSetupAuthorizationAndSessionLifecycle(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CONFIG", dir)
	t.Setenv("DATA", dir)
	t.Setenv("JWT_SECRET", "local-fixture-secret-with-enough-length")
	if _, err := db.Init(); err != nil {
		t.Fatal(err)
	}
	db.Migrate()
	conn, err := db.DB.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })
	if err := service.OpenSessions(dir); err != nil {
		t.Fatal(err)
	}
	gin.SetMode(gin.TestMode)
	router := gin.New()
	api := router.Group("/api")
	RegisterAnonController(api)
	api.Use(AuthMiddleware(true))
	RegisterVehicleController(api)
	RegisterFilesController(api)
	RegisterAuthController(api)
	call := func(method, path, body, token string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		res := httptest.NewRecorder()
		router.ServeHTTP(res, req)
		return res
	}
	first := `{"email":"owner@example.test","password":"test-password-123","name":"Owner","distanceUnit":1,"currency":"EUR"}`
	if response := call(http.MethodPost, "/api/auth/initialize", first, ""); response.Code != 201 {
		t.Fatalf("setup: %d %s", response.Code, response.Body)
	}
	if response := call(http.MethodPost, "/api/auth/initialize", first, ""); response.Code != 409 {
		t.Fatalf("configured setup: %d", response.Code)
	}
	var owner db.User
	if err := db.DB.First(&owner, "email = ?", "owner@example.test").Error; err != nil {
		t.Fatal(err)
	}
	viewer := db.User{Email: "viewer@example.test", Role: db.USER, Name: "Viewer"}
	if err := viewer.SetPassword("viewer-password-123"); err != nil {
		t.Fatal(err)
	}
	if err := db.DB.Create(&viewer).Error; err != nil {
		t.Fatal(err)
	}
	vehicle := db.Vehicle{Nickname: "Private vehicle"}
	if err := db.DB.Create(&vehicle).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.DB.Create(&db.UserVehicle{UserID: owner.ID, VehicleID: vehicle.ID, IsOwner: true}).Error; err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "test.pdf")
	if err := os.WriteFile(path, []byte("test attachment"), 0600); err != nil {
		t.Fatal(err)
	}
	file, err := service.CreateAttachment(path, "test.pdf", 15, "application/pdf", owner.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err := service.CreateVehicleAttachment(vehicle.ID, file.ID, "Invoice"); err != nil {
		t.Fatal(err)
	}

	tokens, err := service.NewSession(&viewer)
	if err != nil {
		t.Fatal(err)
	}
	vehiclePath, filePath := "/api/vehicles/"+vehicle.ID, "/api/attachments/"+file.ID+"/file"
	if response := call(http.MethodGet, vehiclePath, "", tokens.Refresh); response.Code != 401 {
		t.Fatalf("access requires access token: %d", response.Code)
	}
	for _, route := range []string{vehiclePath, filePath} {
		if response := call(http.MethodGet, route, "", tokens.Access); response.Code != 404 {
			t.Fatalf("private resource: %d", response.Code)
		}
	}
	if err := db.DB.Create(&db.UserVehicle{UserID: viewer.ID, VehicleID: vehicle.ID}).Error; err != nil {
		t.Fatal(err)
	}
	for _, route := range []string{vehiclePath, filePath} {
		if response := call(http.MethodGet, route, "", tokens.Access); response.Code != 200 {
			t.Fatalf("shared resource: %d %s", response.Code, response.Body)
		}
	}
	if response := call(http.MethodPut, vehiclePath, "", tokens.Access); response.Code != 404 {
		t.Fatalf("owner-only edit: %d", response.Code)
	}
	_, rotated, err := service.RotateSession(tokens.Refresh)
	if err != nil {
		t.Fatal(err)
	}
	if response := call(http.MethodGet, vehiclePath, "", tokens.Access); response.Code != 401 {
		t.Fatalf("old session after refresh: %d", response.Code)
	}
	if err := db.DB.Model(&viewer).Update("is_disabled", true).Error; err != nil {
		t.Fatal(err)
	}
	if response := call(http.MethodGet, vehiclePath, "", rotated.Access); response.Code != 401 {
		t.Fatalf("disabled account: %d", response.Code)
	}
}
