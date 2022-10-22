package handlers

import (
	"backend/app/auth"
	"backend/app/models"
	"errors"
	"testing"
	"time"

	"gopkg.in/guregu/null.v4"
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

	if joinTableErr := db.SetupJoinTable(&models.User{}, "Datasets", &models.UserDataset{}); joinTableErr != nil {
		t.Fatalf("failed to setup join table: %v", joinTableErr)
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
		t.Fatalf("number of returned datasets should be 2: got %v", len(returnedDatasets))
	}

	// check proper ordering
	if returnedDatasets[0].ID != datasets[1].ID {
		t.Fatal("returned datasets are not ordered by CreatedAt")
	}

	if returnedDatasets[1].ID != datasets[0].ID {
		t.Fatal("returned datasets are not ordered by CreatedAt")
	}
}

func TestDatasetStats(t *testing.T) {
	db, cleanup := setupDBForDatasetsHandlerTests(t)
	defer cleanup()

	dataset := &models.Dataset{
		Name: "dataset1",
		Type: models.EntityAnnotation,
	}

	if result := db.Create(&dataset); result.Error != nil {
		t.Fatalf("failed to create dataset: %v", result.Error)
	}

	samples := []models.Sample{
		{DatasetID: dataset.ID, Status: models.Accepted.ToNullString()},
		{DatasetID: dataset.ID, AssignedTo: null.IntFrom(1)},
		{DatasetID: dataset.ID},
	}

	if result := db.Create(&samples); result.Error != nil {
		t.Fatalf("failed to create samples: %v", result.Error)
	}

	returnedDatasets := NewDatasetsHandler(db).GetDatasets()
	if len(returnedDatasets) != 1 {
		t.Fatalf("number of returned datasets should be 1: got %v", len(returnedDatasets))
	}

	// check stats
	stats := returnedDatasets[0].Stats
	if stats.TotalSamples != int64(len(samples)) {
		t.Fatalf("incorrect number of total samples: got %v, expected: %v", stats.TotalSamples, len(samples))
	}

	if stats.CompletedSamples != 1 {
		t.Fatalf("number of completed samples should be 1: %v", stats.CompletedSamples)
	}

	if stats.PendingSamples != 1 {
		t.Fatalf("number of pending samples should be 1: %v", stats.PendingSamples)
	}
}

func TestGetDataset(t *testing.T) {
	db, cleanup := setupDBForDatasetsHandlerTests(t)
	defer cleanup()

	datasets := []models.Dataset{
		{Name: "dataset1", Type: models.EntityAnnotation},
		{Name: "dataset2", Type: models.EntityAnnotation},
	}

	if result := db.Create(&datasets); result.Error != nil {
		t.Fatalf("failed to create datasets: %v", result.Error)
	}

	returnedDataset, datasetErr := NewDatasetsHandler(db).GetDataset(datasets[1].ID)
	if datasetErr != nil {
		t.Fatalf("unexpected error fetching dataset: %v", datasetErr)
	}

	if returnedDataset.ID != datasets[1].ID {
		t.Fatal("returned wrong dataset")
	}
}

func TestDeleteDataset(t *testing.T) {
	db, cleanup := setupDBForDatasetsHandlerTests(t)
	defer cleanup()

	dataset := &models.Dataset{
		Name: "dataset1",
		Type: models.EntityAnnotation,
	}

	if result := db.Create(&dataset); result.Error != nil {
		t.Fatalf("failed to create dataset: %v", result.Error)
	}

	if deleteErr := NewDatasetsHandler(db).DeleteDataset(dataset.ID); deleteErr != nil {
		t.Fatalf("unexpected error occurred while deleting dataset: %v", deleteErr)
	}

	datasetResult := &models.Dataset{}
	if result := db.First(&datasetResult, dataset.ID); !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		t.Fatalf("unexpected error occurred when fetching non-existent dataset: %v", result.Error)
	}
}

func TestGetDatasetData(t *testing.T) {
	db, cleanup := setupDBForDatasetsHandlerTests(t)
	defer cleanup()

	dataset := &models.Dataset{
		Name: "dataset1",
		Type: models.EntityAnnotation,
	}

	if result := db.Create(&dataset); result.Error != nil {
		t.Fatalf("failed to create dataset: %v", result.Error)
	}

	data, err := NewDatasetsHandler(db).GetDatasetData(dataset.ID)
	if err != nil {
		t.Fatalf("unexpected error fetching dataset data: %v", err)
	}

	if data.ID != dataset.ID {
		t.Fatalf("dataset ids do not match: got %v, expected %v", data.ID, dataset.ID)
	}

	if data.Name != dataset.Name {
		t.Fatalf("dataset names do not match: got %v, expected %v", data.Name, dataset.Name)
	}

	if data.Type != dataset.Type {
		t.Fatalf("dataset types do not match: got %v, expected %v", data.Type, dataset.Type)
	}
}

func TestGetDatasetForUser(t *testing.T) {
	db, cleanup := setupDBForDatasetsHandlerTests(t)
	defer cleanup()

	userAuth := auth.NewUserAuth(db)
	user1, user1Err := userAuth.CreateUser("email@email.com", "password", models.AdminRole)
	if user1Err != nil {
		t.Fatalf("failed to create user: %v", user1Err)
	}

	user2, user2Err := userAuth.CreateUser("email2@email.com", "password", models.AdminRole)
	if user2Err != nil {
		t.Fatalf("failed to create user: %v", user2Err)
	}

	datasets := []models.Dataset{
		{Name: "dataset1", Type: models.EntityAnnotation},
		{Name: "dataset1", Type: models.EntityAnnotation},
		{Name: "dataset3", Type: models.EntityAnnotation},
	}

	if result := db.Create(&datasets); result.Error != nil {
		t.Fatalf("failed to create datasets: %v", result.Error)
	}

	permsHandler := NewUserDatasetPermsHandler(db)
	if err := permsHandler.AddDatasetToUserPerms(user1.ID, datasets[0].ID); err != nil {
		t.Fatalf("failed to assign permissions: %v", err)
	}

	if err := permsHandler.AddDatasetToUserPerms(user1.ID, datasets[1].ID); err != nil {
		t.Fatalf("failed to assign permissions: %v", err)
	}

	if err := permsHandler.AddDatasetToUserPerms(user2.ID, datasets[2].ID); err != nil {
		t.Fatalf("failed to assign permissions: %v", err)
	}

	data, err := NewDatasetsHandler(db).GetDatasetsForUser(user1)
	if err != nil {
		t.Fatalf("unexpected error fetching datasets for user: %v", err)
	}

	for _, dataset := range data {
		if dataset.ID == datasets[2].ID {
			t.Fatalf("handler returned dataset that was not assigned to the given user")
		}
	}
}
