package controllers

import (
	"backend/app/auth"
	"backend/app/models"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/matryer/is"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupDBForUsersControllerTests(t *testing.T) (*gorm.DB, func() error) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})

	if err != nil {
		t.Fatalf("failed to open sqlite connection: %v", err)
	}

	if migrationErr := db.AutoMigrate(&models.User{}); migrationErr != nil {
		t.Fatalf("failed to migrate user: %v", migrationErr)
	}

	if migrationErr := db.AutoMigrate(&models.AuthToken{}); migrationErr != nil {
		t.Fatalf("failed to migrate auth token: %v", migrationErr)
	}

	if migrationErr := db.AutoMigrate(&models.RefreshToken{}); migrationErr != nil {
		t.Fatalf("failed to migrate refresh token: %v", migrationErr)
	}

	sqlDB, sqlErr := db.DB()
	if sqlErr != nil {
		t.Fatalf("failed to obtain SQL DB: %v", sqlErr)
	}
	return db, sqlDB.Close
}

func setupUsersController(t *testing.T) (*gorm.DB, func() error, *mux.Router) {
	db, cleanup := setupDBForUsersControllerTests(t)
	tokenAuth := auth.NewTokenAuth(db)
	router := mux.NewRouter()
	authController := NewUsersController(tokenAuth)
	authController.Init(router)
	return db, cleanup, router
}

func TestGetUserWithoutAutth(t *testing.T) {
	_, cleanup, router := setupUsersController(t)
	defer cleanup()
	is := is.New(t)

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	is.Equal(rr.Code, http.StatusUnauthorized)
}

func TestGetUser(t *testing.T) {
	db, cleanup, router := setupUsersController(t)
	defer cleanup()
	is := is.New(t)

	email := "user@email.com"
	user := models.User{Email: email, Role: models.AdminRole}
	is.NoErr(db.Create(&user).Error)

	tokenAuth := auth.NewTokenAuth(db)
	authToken, tokenErr := tokenAuth.CreateAuthToken(&user)
	is.NoErr(tokenErr)
	authCookie, _ := tokenAuth.CreateAuthCookies(authToken)

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(authCookie)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	is.Equal(rr.Code, http.StatusOK)

	responseUser := &models.User{}
	is.NoErr(json.NewDecoder(rr.Body).Decode(responseUser))
	is.Equal(user.ID, responseUser.ID)
	is.Equal(responseUser.Email, email)
}
