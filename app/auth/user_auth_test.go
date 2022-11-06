package auth

import (
	"backend/app/models"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupDBForUserTests(t *testing.T) (*gorm.DB, func() error) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})

	if err != nil {
		t.Fatalf("failed to open sqlite connection: %v", err)
	}

	if migrationErr := db.AutoMigrate(&models.User{}); migrationErr != nil {
		t.Fatalf("failed to migrate user: %v", migrationErr)
	}

	sqlDB, sqlErr := db.DB()
	if sqlErr != nil {
		t.Fatalf("failed to obtain SQL DB: %v", sqlErr)
	}
	return db, sqlDB.Close
}

func TestCreateUser(t *testing.T) {
	db, cleanup := setupDBForUserTests(t)
	defer cleanup()

	userAuth := NewUserAuth(db)
	email := "email@email.com"
	role := models.AdminRole

	var userCount int64 = 0
	db.Model(&models.User{}).Count(&userCount)

	if userCount != 0 {
		t.Fatalf("the number of users should be 0")
	}

	createdUser, userErr := userAuth.CreateUser(email, "pass", role)
	if userErr != nil {
		t.Fatalf("failed to create user: %v", userErr)
	}

	db.Model(&models.User{}).Count(&userCount)
	if userCount != 1 {
		t.Fatalf("the number of users should be 1")
	}

	dbUser := &models.User{}
	if result := db.First(dbUser); result.Error != nil {
		t.Fatalf("failed to fetch user from DB: %v", result.Error)
	}

	if createdUser.Email != email {
		t.Fatalf("returned user has incorrect email: got %v, expected %v", createdUser.Email, email)
	}

	if dbUser.Email != email {
		t.Fatalf("stored user has incorrect email: got %v, expected %v", createdUser.Email, email)
	}
}

func TestCheckUserPassword(t *testing.T) {
	db, cleanup := setupDBForUserTests(t)
	defer cleanup()

	userAuth := NewUserAuth(db)
	email := "email@email.com"
	pass := "pass"
	role := models.AdminRole

	createdUser, createErr := userAuth.CreateUser(email, pass, role)
	if createErr != nil {
		t.Fatalf("failed to create user: %v", createErr)
	}

	user, userErr := userAuth.CheckUserPassword(email, pass)
	if userErr != nil {
		t.Fatalf("CheckUserPassword returned unexpected error: %v", userErr)
	}

	if user.ID != createdUser.ID {
		t.Fatalf("CheckUserPassword returned wrong user: got %v, expected %v", createdUser.ID, user.ID)
	}
}

func TestCheckUserPasswordWithIncorrectEmail(t *testing.T) {
	db, cleanup := setupDBForUserTests(t)
	defer cleanup()

	userAuth := NewUserAuth(db)
	email := "email@email.com"
	pass := "pass"
	role := models.AdminRole

	_, createErr := userAuth.CreateUser(email, pass, role)
	if createErr != nil {
		t.Fatalf("failed to create user: %v", createErr)
	}

	_, userErr := userAuth.CheckUserPassword(email+"random", pass)
	if userErr != ErrWrongEmailOrPassword {
		t.Fatalf("CheckUserPassword returned unexpected error: %v", userErr)
	}
}

func TestCheckUserPasswordWithIncorrectPassword(t *testing.T) {
	db, cleanup := setupDBForUserTests(t)
	defer cleanup()

	userAuth := NewUserAuth(db)
	email := "email@email.com"
	pass := "pass"
	role := models.AdminRole

	_, createErr := userAuth.CreateUser(email, pass, role)
	if createErr != nil {
		t.Fatalf("failed to create user: %v", createErr)
	}

	_, userErr := userAuth.CheckUserPassword(email, pass+"random")
	if userErr != ErrWrongEmailOrPassword {
		t.Fatalf("CheckUserPassword returned unexpected error: %v", userErr)
	}
}
