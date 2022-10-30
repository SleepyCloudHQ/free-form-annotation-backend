package handlers

import (
	"backend/app/auth"
	"backend/app/models"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupDBForAuthHandlerTests(t *testing.T) (*gorm.DB, func() error) {
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

func TestFailing(t *testing.T) {
	t.Fatal("fail")
}

func TestRegister(t *testing.T) {
	db, cleanup := setupDBForAuthHandlerTests(t)
	defer cleanup()

	handler := NewAuthHandler(auth.NewUserAuth(db), auth.NewTokenAuth(db))
	user, cookies, err := handler.Register("email@email.com", "pass")
	if err != nil {
		t.Fatalf("unexpected error occurred while registering: %v", err)
	}

	authToken := &models.AuthToken{}
	if result := db.Preload("RefreshToken").Where("user_id = ?", user.ID).First(&authToken); result.Error != nil {
		t.Fatalf("unexpected error occurred while fetching auth token: %v", result.Error)
	}

	if authToken.Token != cookies.AuthTokenCookie.Value {
		t.Fatalf("auth cookie value is incorrect: got %v, expected %v", cookies.AuthTokenCookie.Value, authToken.Token)
	}

	if authToken.RefreshToken.Token != cookies.RefreshTokenCookie.Value {
		t.Fatalf("refresh token cookie value is incorrect: got %v, expected %v", cookies.RefreshTokenCookie.Value, authToken.RefreshToken.Token)
	}
}

func TestLogin(t *testing.T) {
	db, cleanup := setupDBForAuthHandlerTests(t)
	defer cleanup()

	userAuth := auth.NewUserAuth(db)
	handler := NewAuthHandler(userAuth, auth.NewTokenAuth(db))

	email := "email@email.com"
	pass := "pass"

	if _, userErr := userAuth.CreateUser(email, pass, models.AnnotatorRole); userErr != nil {
		t.Fatalf("failed to create user: %v", userErr)
	}

	user, cookies, err := handler.Login("email@email.com", "pass")
	if err != nil {
		t.Fatalf("unexpected error occurred while registering: %v", err)
	}

	authToken := &models.AuthToken{}
	if result := db.Preload("RefreshToken").Where("user_id = ?", user.ID).First(&authToken); result.Error != nil {
		t.Fatalf("unexpected error occurred while fetching auth token: %v", result.Error)
	}

	if authToken.Token != cookies.AuthTokenCookie.Value {
		t.Fatalf("auth cookie value is incorrect: got %v, expected %v", cookies.AuthTokenCookie.Value, authToken.Token)
	}

	if authToken.RefreshToken.Token != cookies.RefreshTokenCookie.Value {
		t.Fatalf("refresh token cookie value is incorrect: got %v, expected %v", cookies.RefreshTokenCookie.Value, authToken.RefreshToken.Token)
	}
}
