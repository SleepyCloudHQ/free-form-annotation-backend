package handlers

import (
	"backend/app/models"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupDBForUserDatasetPermsTests(t *testing.T) (*gorm.DB, func() error) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})

	if err != nil {
		t.Fatalf("failed to open sqlite connection: %v", err)
	}

	if migrationErr := db.AutoMigrate(&models.UserDataset{}); migrationErr != nil {
		t.Fatalf("failed to migrate user datasets: %v", migrationErr)
	}

	sqlDB, sqlErr := db.DB()
	if sqlErr != nil {
		t.Fatalf("failed to obtain SQL DB: %v", sqlErr)
	}
	return db, sqlDB.Close
}

func TestAddPerms(t *testing.T) {
	db, cleanup := setupDBForUserDatasetPermsTests(t)
	defer cleanup()

	handler := NewUserDatasetPermsHandler(db)
	var userId uint = 1
	var datasetId uint = 2

	if err := handler.AddDatasetToUserPerms(userId, datasetId); err != nil {
		t.Fatalf("unexpected error occurred while adding permissions: %v", err)
	}

	var count int64
	if err := db.Model(&models.UserDataset{}).Count(&count).Error; err != nil {
		t.Fatalf("unexpected error occurred while getting model count: %v", err)
	}

	if count != 1 {
		t.Fatalf("incorred number of records: got %v, expected: 1", count)
	}

	perm := &models.UserDataset{}
	if err := db.First(&perm).Error; err != nil {
		t.Fatalf("unexpected error while fetching perm record: %v", err)
	}

	if perm.DatasetID != datasetId {
		t.Fatalf("dataset ids do not match: got %v, expected: %v", perm.DatasetID, datasetId)
	}

	if perm.UserID != userId {
		t.Fatalf("user ids do not match: got %v, expected: %v", perm.UserID, userId)
	}
}

func TestDeletePerms(t *testing.T) {
	db, cleanup := setupDBForUserDatasetPermsTests(t)
	defer cleanup()

	handler := NewUserDatasetPermsHandler(db)
	var userId uint = 1
	var datasetId1 uint = 1
	var datasetId2 uint = 2

	if err := handler.AddDatasetToUserPerms(userId, datasetId1); err != nil {
		t.Fatalf("unexpected error occurred while adding permissions: %v", err)
	}

	if err := handler.AddDatasetToUserPerms(userId, datasetId2); err != nil {
		t.Fatalf("unexpected error occurred while adding permissions: %v", err)
	}

	var count int64
	if err := db.Model(&models.UserDataset{}).Count(&count).Error; err != nil {
		t.Fatalf("unexpected error occurred while getting model count: %v", err)
	}

	if count != 2 {
		t.Fatalf("incorred number of records: got %v, expected: 2", count)
	}

	if err := handler.DeleteDatasetToUserPerms(userId, datasetId2); err != nil {
		t.Fatalf("unexpected error occurred while deleting permissions: %v", err)
	}

	if err := db.Model(&models.UserDataset{}).Count(&count).Error; err != nil {
		t.Fatalf("unexpected error occurred while getting model count: %v", err)
	}

	if count != 1 {
		t.Fatalf("incorred number of records: got %v, expected: 1", count)
	}

	// fetch the remaining permission and check the values
	perm := &models.UserDataset{}
	if err := db.First(&perm).Error; err != nil {
		t.Fatalf("unexpected error while fetching perm record: %v", err)
	}

	if perm.DatasetID != datasetId1 {
		t.Fatalf("dataset ids do not match: got %v, expected: %v", perm.DatasetID, datasetId1)
	}

	if perm.UserID != userId {
		t.Fatalf("user ids do not match: got %v, expected: %v", perm.UserID, userId)
	}
}
