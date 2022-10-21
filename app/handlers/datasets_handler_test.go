package handlers

import (
	"backend/app/models"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupDBForDatasetsHandlerTests(t *testing.T) (*gorm.DB, func() error) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})

	if err != nil {
		t.Fatalf("failed to open sqlite connection: %v", err)
	}

	if migrationErr := db.AutoMigrate(&models.User{}); migrationErr != nil {
		t.Fatalf("failed to migrate user: %v", migrationErr)
	}

	if migrationErr := db.AutoMigrate(&models.Sample{}); migrationErr != nil {
		t.Fatalf("failed to migrate sample: %v", migrationErr)
	}

	if migrationErr := db.AutoMigrate(&models.Dataset{}); migrationErr != nil {
		t.Fatalf("failed to migrate dataset: %v", migrationErr)
	}

	if migrationErr := db.AutoMigrate(&models.UserDataset{}); migrationErr != nil {
		t.Fatalf("failed to create user dataset pivot table: %v", migrationErr)
	}

	sqlDB, sqlErr := db.DB()
	if sqlErr != nil {
		t.Fatalf("failed to obtain SQL DB: %v", sqlErr)
	}
	return db, sqlDB.Close
}

func TestGetDatasets(t *testing.T) {
	db, cleanup := setupDBForDatasetsHandlerTests(t)
	defer cleanup()

	datasets := []models.Dataset{
		{Name: "dataset1", Type: models.EntityAnnotation, Model: gorm.Model{CreatedAt: time.Now().Add(-time.Hour)}},
		{Name: "dataset2", Type: models.EntityAnnotation, Model: gorm.Model{CreatedAt: time.Now()}},
	}

	if result := db.Create(&datasets); result.Error != nil {
		t.Fatalf("failed to create datasets: %v", result.Error)
	}

	returnedDatasets := NewDatasetsHandler(db).GetDatasets()
	if len(returnedDatasets) != 2 {
		t.Fatalf("number of returned datasets is not 2: got %v", len(returnedDatasets))
	}

	// check proper ordering
	if returnedDatasets[0].ID != datasets[1].ID {
		t.Fatal("returned datasets are not ordered by CreatedAt")
	}

	if returnedDatasets[1].ID != datasets[0].ID {
		t.Fatal("returned datasets are not ordered by CreatedAt")
	}
}
