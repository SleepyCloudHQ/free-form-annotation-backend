package auth

import (
	"backend/app/models"
	"errors"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupDBForTokenTests(t *testing.T) (*gorm.DB, func() error) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})

	if err != nil {
		t.Fatalf("failed to open sqlite connection: %v", err)
	}

	if migrationErr := db.AutoMigrate(&models.User{}); migrationErr != nil {
		t.Fatalf("failed to migrate user: %v", err)
	}

	if migrationErr := db.AutoMigrate(&models.AuthToken{}); migrationErr != nil {
		t.Fatalf("failed to migrate auth token: %v", err)
	}

	if migrationErr := db.AutoMigrate(&models.RefreshToken{}); migrationErr != nil {
		t.Fatalf("failed to migrate refresh token: %v", err)
	}

	sqlDB, sqlErr := db.DB()
	if sqlErr != nil {
		t.Fatalf("failed to obtain SQL DB: %v", err)
	}
	return db, sqlDB.Close
}

func TestCreateAuthToken(t *testing.T) {
	db, cleanup := setupDBForTokenTests(t)
	defer cleanup()

	tokenAuth := NewTokenAuth(db)

	userAuth := NewUserAuth(db)
	user, userErr := userAuth.CreateUser("email@email.com", "password", models.AdminRole)
	if userErr != nil {
		t.Fatalf("failed to create user: %v", userErr)
	}

	var authTokenCount int64 = 0
	db.Model(&models.AuthToken{}).Count(&authTokenCount)

	var refreshTokenCount int64 = 0
	db.Model(&models.RefreshToken{}).Count(&refreshTokenCount)

	if authTokenCount != 0 {
		t.Fatalf("the number of auth tokens should be 0")
	}

	if refreshTokenCount != 0 {
		t.Fatalf("the number of refresh tokens should be 0")
	}

	authToken, tokenErr := tokenAuth.CreateAuthToken(user)
	if tokenErr != nil {
		t.Fatalf("failed to create auth token: %v", tokenErr)
	}

	db.Model(&models.AuthToken{}).Count(&authTokenCount)
	db.Model(&models.RefreshToken{}).Count(&refreshTokenCount)

	if authTokenCount != 1 {
		t.Fatalf("the number of auth tokens should be 1")
	}

	if refreshTokenCount != 1 {
		t.Fatalf("the number of refresh tokens should be 1")
	}

	fetchedToken := &models.AuthToken{}
	if result := db.First(fetchedToken); result.Error != nil {
		t.Fatalf("failed to fetch the created auth token from db: %v", result.Error)
	}

	if authToken.ID != fetchedToken.ID {
		t.Fatalf("auth token ids do not match")
	}

	if authToken.UserID != user.ID {
		t.Fatalf("user ids do not match")
	}
}

func TestRefreshToken(t *testing.T) {
	db, cleanup := setupDBForTokenTests(t)
	defer cleanup()

	tokenAuth := NewTokenAuth(db)

	userAuth := NewUserAuth(db)
	user, userErr := userAuth.CreateUser("email@email.com", "password", models.AdminRole)
	if userErr != nil {
		t.Fatalf("failed to create user: %v", userErr)
	}

	authToken, tokenErr := tokenAuth.CreateAuthToken(user)
	if tokenErr != nil {
		t.Fatalf("failed to create auth token: %v", tokenErr)
	}

	_, refreshErr := tokenAuth.RefreshToken(authToken.RefreshToken.Token)
	if refreshErr != nil {
		t.Fatalf("failed to refresh auth token: %v", refreshErr)
	}

	// check if old tokens were deleted
	if result := db.First(&models.AuthToken{}, authToken.ID); !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		t.Fatalf("Record not deleted or other error occurred: %v", result.Error)
	}

	if result := db.First(&models.RefreshToken{}, authToken.RefreshToken.ID); !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		t.Fatalf("Record not deleted or other error occurred: %v", result.Error)
	}
}

func TestCheckAuthToken(t *testing.T) {
	db, cleanup := setupDBForTokenTests(t)
	defer cleanup()
	tokenAuth := NewTokenAuth(db)

	userAuth := NewUserAuth(db)
	user, userErr := userAuth.CreateUser("email@email.com", "password", models.AdminRole)
	if userErr != nil {
		t.Fatalf("failed to create user: %v", userErr)
	}

	authToken, tokenErr := tokenAuth.CreateAuthToken(user)
	if tokenErr != nil {
		t.Fatalf("failed to create auth token: %v", tokenErr)
	}

	authenticatedUser, authErr := tokenAuth.CheckAuthToken(authToken.Token)
	if authErr != nil {
		t.Fatalf("failed to check token: %v", authErr)
	}

	if authenticatedUser.ID != user.ID {
		t.Fatalf("user ids are not matching")
	}
}

func TestCheckIncorrectAuthToken(t *testing.T) {
	db, cleanup := setupDBForTokenTests(t)
	defer cleanup()
	tokenAuth := NewTokenAuth(db)

	_, authErr := tokenAuth.CheckAuthToken("random token")
	if !errors.Is(authErr, ErrInvalidToken) {
		t.Fatalf("unexpected error returned: %v", authErr)
	}
}

func TestCheckExpiredAuthToken(t *testing.T) {
	db, cleanup := setupDBForTokenTests(t)
	defer cleanup()
	tokenAuth := NewTokenAuth(db)

	userAuth := NewUserAuth(db)
	user, userErr := userAuth.CreateUser("email@email.com", "password", models.AdminRole)
	if userErr != nil {
		t.Fatalf("failed to create user: %v", userErr)
	}

	authToken, tokenErr := tokenAuth.CreateAuthToken(user)
	if tokenErr != nil {
		t.Fatalf("failed to create auth token: %v", tokenErr)
	}

	if result := db.Model(authToken).Update("expires_at", time.Now().Add(-time.Hour)); result.Error != nil {
		t.Fatalf("failed to update auth token expiration time: %v", tokenErr)
	}

	_, authErr := tokenAuth.CheckAuthToken(authToken.Token)
	if !errors.Is(authErr, ErrTokenExpired) {
		t.Fatalf("unexpected error returned: %v", authErr)
	}
}
