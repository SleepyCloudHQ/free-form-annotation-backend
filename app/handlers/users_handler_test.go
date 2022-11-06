package handlers

import (
	"backend/app/auth"
	"backend/app/models"
	"testing"

	"github.com/matryer/is"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupDBForUsersHandlerTests(t *testing.T) (*gorm.DB, func() error) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})

	if err != nil {
		t.Fatalf("failed to open sqlite connection: %v", err)
	}

	if migrationErr := db.AutoMigrate(&models.User{}); migrationErr != nil {
		t.Fatalf("failed to migrate user: %v", migrationErr)
	}

	if migrationErr := db.AutoMigrate(&models.Dataset{}); migrationErr != nil {
		t.Fatalf("failed to migrate dataset: %v", migrationErr)
	}

	if migrationErr := db.AutoMigrate(&models.UserDataset{}); migrationErr != nil {
		t.Fatalf("failed to migrate user datasets: %v", migrationErr)
	}

	if joinTableErr := db.SetupJoinTable(&models.User{}, "Datasets", &models.UserDataset{}); joinTableErr != nil {
		t.Fatalf("failed to setup join table: %v", joinTableErr)
	}

	sqlDB, sqlErr := db.DB()
	if sqlErr != nil {
		t.Fatalf("failed to obtain SQL DB: %v", sqlErr)
	}
	return db, sqlDB.Close
}

func TestGetUsers(t *testing.T) {
	is := is.New(t)
	db, cleanup := setupDBForUsersHandlerTests(t)
	defer cleanup()

	userAuth := auth.NewUserAuth(db)
	handler := NewUsersHandler(db)

	user1, user1Err := userAuth.CreateUser("email1@email.com", "pass", models.AdminRole)
	is.NoErr(user1Err)

	user2, user2Err := userAuth.CreateUser("email2@email.com", "pass", models.AdminRole)
	is.NoErr(user2Err)

	users := handler.GetUsers()
	is.Equal(len(users), 2)
	is.Equal(user1.ID, users[0].ID)
	is.Equal(user2.ID, users[1].ID)
}

func TestGetUsersWithDatasets(t *testing.T) {
	is := is.New(t)
	db, cleanup := setupDBForUsersHandlerTests(t)
	defer cleanup()

	userAuth := auth.NewUserAuth(db)
	permsHandler := NewUserDatasetPermsHandler(db)
	handler := NewUsersHandler(db)

	user, userErr := userAuth.CreateUser("email1@email.com", "pass", models.AdminRole)
	is.NoErr(userErr)

	datasets := []models.Dataset{
		{Name: "dataset1", Type: models.EntityAnnotation},
		{Name: "dataset2", Type: models.EntityAnnotation},
		{Name: "dataset3", Type: models.EntityAnnotation},
	}

	result := db.Create(&datasets)
	is.NoErr(result.Error)

	permsHandler.AddDatasetToUserPerms(user.ID, datasets[0].ID)
	permsHandler.AddDatasetToUserPerms(user.ID, datasets[1].ID)

	usersWithDatasets := handler.GetUsersWithDatasets()
	is.Equal(len(usersWithDatasets), 1)

	returnedDatasets := usersWithDatasets[0].Datasets
	is.Equal(len(returnedDatasets), 2)
	is.Equal(returnedDatasets[0].ID, datasets[0].ID)
	is.Equal(returnedDatasets[1].ID, datasets[1].ID)
}

func TestPatchUserRole(t *testing.T) {
	is := is.New(t)
	db, cleanup := setupDBForUsersHandlerTests(t)
	defer cleanup()

	handler := NewUsersHandler(db)
	userAuth := auth.NewUserAuth(db)
	user, userErr := userAuth.CreateUser("email1@email.com", "pass", models.AdminRole)
	is.NoErr(userErr)

	newRole := models.AnnotatorRole
	_, patchErr := handler.PatchUserRole(user.ID, newRole)
	is.NoErr(patchErr)

	updatedUser := &models.User{}
	is.NoErr(db.First(updatedUser, user.ID).Error)
	is.Equal(updatedUser.Role, newRole)
}
