package handlers

import (
	"backend/app/models"
	"testing"

	"github.com/matryer/is"
	"gopkg.in/guregu/null.v4"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupDBForSamplesHandlerTests(t *testing.T) (*gorm.DB, func() error) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})

	if err != nil {
		t.Fatalf("failed to open sqlite connection: %v", err)
	}

	if migrationErr := db.AutoMigrate(&models.Sample{}); migrationErr != nil {
		t.Fatalf("failed to migrate sample: %v", migrationErr)
	}

	if migrationErr := db.AutoMigrate(&models.Dataset{}); migrationErr != nil {
		t.Fatalf("failed to migrate dataset: %v", migrationErr)
	}

	sqlDB, sqlErr := db.DB()
	if sqlErr != nil {
		t.Fatalf("failed to obtain SQL DB: %v", sqlErr)
	}
	return db, sqlDB.Close
}

func TestGetSamples(t *testing.T) {
	db, cleanup := setupDBForSamplesHandlerTests(t)
	defer cleanup()

	is := is.New(t)
	handler := NewSamplesHandler(db)

	dataset1Samples := []models.Sample{
		{Text: "Sample text"},
		{Text: "Sample text"},
		{Text: "Sample text"},
	}

	dataset2Samples := []models.Sample{
		{Text: "Sample text"},
		{Text: "Sample text"},
	}

	datasets := []models.Dataset{
		{Name: "dataset1", Type: models.EntityAnnotation, Samples: dataset1Samples},
		{Name: "dataset2", Type: models.EntityAnnotation, Samples: dataset2Samples},
	}

	is.NoErr(db.Create(&datasets).Error)

	samples, samplesErr := handler.GetSamples(datasets[1].ID)
	is.NoErr(samplesErr)
	is.Equal(len(samples), 2)

	is.Equal(samples[0].ID, datasets[1].Samples[0].ID)
	is.Equal(samples[1].ID, datasets[1].Samples[1].ID)
}

func TestGetSamplesWithStatus(t *testing.T) {
	db, cleanup := setupDBForSamplesHandlerTests(t)
	defer cleanup()

	is := is.New(t)
	handler := NewSamplesHandler(db)

	dataset1Samples := []models.Sample{
		{Text: "Sample text", Status: models.Accepted.ToNullString()},
		{Text: "Sample text", Status: models.Rejected.ToNullString()},
		{Text: "Sample text"},
	}

	dataset2Samples := []models.Sample{
		{Text: "Sample text", Status: models.Accepted.ToNullString()},
	}

	datasets := []models.Dataset{
		{Name: "dataset1", Type: models.EntityAnnotation, Samples: dataset1Samples},
		{Name: "dataset2", Type: models.EntityAnnotation, Samples: dataset2Samples},
	}

	is.NoErr(db.Create(&datasets).Error)

	samples, samplesErr := handler.GetSamplesWithStatus(datasets[0].ID, models.Accepted)
	is.NoErr(samplesErr)
	is.Equal(len(samples), 1)

	is.Equal(samples[0].ID, datasets[0].Samples[0].ID)

	// no samples with requested status
	samples, samplesErr = handler.GetSamplesWithStatus(datasets[0].ID, models.Uncertain)
	is.NoErr(samplesErr)
	is.Equal(len(samples), 0)
}

func TestGetSample(t *testing.T) {
	db, cleanup := setupDBForSamplesHandlerTests(t)
	defer cleanup()

	is := is.New(t)
	handler := NewSamplesHandler(db)

	dataset1Samples := []models.Sample{
		{Text: "Sample text"},
	}

	dataset2Samples := []models.Sample{
		{Text: "Sample text"},
	}

	datasets := []models.Dataset{
		{Name: "dataset1", Type: models.EntityAnnotation, Samples: dataset1Samples},
		{Name: "dataset2", Type: models.EntityAnnotation, Samples: dataset2Samples},
	}

	is.NoErr(db.Create(&datasets).Error)

	sample, sampleErr := handler.GetSample(datasets[0].ID, datasets[0].Samples[0].ID)
	is.NoErr(sampleErr)
	is.Equal(sample.ID, datasets[0].Samples[0].ID)
}

func TestGetNonExistentSample(t *testing.T) {
	db, cleanup := setupDBForSamplesHandlerTests(t)
	defer cleanup()

	is := is.New(t)
	handler := NewSamplesHandler(db)

	dataset1Samples := []models.Sample{
		{Text: "Sample text"},
	}

	dataset := &models.Dataset{
		Name:    "dataset1",
		Type:    models.EntityAnnotation,
		Samples: dataset1Samples,
	}

	is.NoErr(db.Create(&dataset).Error)

	_, sampleErr := handler.GetSample(dataset.ID, 0)
	is.Equal(sampleErr, gorm.ErrRecordNotFound)
}

func TestPatchSample(t *testing.T) {
	db, cleanup := setupDBForSamplesHandlerTests(t)
	defer cleanup()

	is := is.New(t)
	handler := NewSamplesHandler(db)

	originalStatus := models.Accepted.ToNullString()
	dataset1Samples := []models.Sample{
		{Text: "Sample text", Status: originalStatus},
	}

	dataset2Samples := []models.Sample{
		{Text: "Sample text", Status: originalStatus},
	}

	datasets := []models.Dataset{
		{Name: "dataset1", Type: models.EntityAnnotation, Samples: dataset1Samples},
		{Name: "dataset2", Type: models.EntityAnnotation, Samples: dataset2Samples},
	}

	is.NoErr(db.Create(&datasets).Error)

	data := &UpdateSampleData{
		Status: models.Rejected.ToNullString(),
	}

	_, updateErr := handler.PatchSample(datasets[0].ID, datasets[0].Samples[0].ID, data)
	is.NoErr(updateErr)

	refreshedSample, sampleErr := handler.GetSample(datasets[0].ID, datasets[0].Samples[0].ID)
	is.NoErr(sampleErr)

	is.Equal(refreshedSample.Status, data.Status)

	otherRefreshedSample, otherSampleErr := handler.GetSample(datasets[1].ID, datasets[1].Samples[0].ID)
	is.NoErr(otherSampleErr)
	is.Equal(otherRefreshedSample.Status, originalStatus)
}

func TestPatchNonExistentSample(t *testing.T) {
	db, cleanup := setupDBForSamplesHandlerTests(t)
	defer cleanup()

	is := is.New(t)
	handler := NewSamplesHandler(db)

	originalStatus := models.Accepted.ToNullString()
	dataset1Samples := []models.Sample{
		{Text: "Sample text", Status: originalStatus},
	}

	dataset := &models.Dataset{
		Name:    "dataset1",
		Type:    models.EntityAnnotation,
		Samples: dataset1Samples,
	}

	is.NoErr(db.Create(&dataset).Error)

	data := &UpdateSampleData{
		Status: models.Rejected.ToNullString(),
	}

	_, updateErr := handler.PatchSample(dataset.ID, 0, data)
	is.Equal(updateErr, gorm.ErrRecordNotFound)
}

func TestFindingAssignedSample(t *testing.T) {
	db, cleanup := setupDBForSamplesHandlerTests(t)
	defer cleanup()

	is := is.New(t)
	handler := NewSamplesHandler(db)

	var userID uint = 0
	dataset1Samples := []models.Sample{
		{Text: "Sample text"},
		{Text: "Sample text", AssignedTo: null.IntFrom(int64(userID))},
	}

	dataset := &models.Dataset{
		Name:    "dataset1",
		Type:    models.EntityAnnotation,
		Samples: dataset1Samples,
	}

	is.NoErr(db.Create(&dataset).Error)
	assignedSample, sampleErr := handler.AssignNextSample(dataset.ID, userID)
	is.NoErr(sampleErr)
	is.Equal(assignedSample.ID, dataset1Samples[1].ID)
}

func TestAssigningNewSample(t *testing.T) {

	db, cleanup := setupDBForSamplesHandlerTests(t)
	defer cleanup()

	is := is.New(t)
	handler := NewSamplesHandler(db)

	dataset1Samples := []models.Sample{
		{Text: "Sample text", Status: models.Accepted.ToNullString()},
		{Text: "Sample text", AssignedTo: null.IntFrom(5)},
		{Text: "Sample text", Status: models.Accepted.ToNullString(), AssignedTo: null.IntFrom(4)},
		{Text: "Sample text"},
		{Text: "Sample text"},
	}

	dataset := &models.Dataset{
		Name:    "dataset1",
		Type:    models.EntityAnnotation,
		Samples: dataset1Samples,
	}

	is.NoErr(db.Create(&dataset).Error)
	assignedSample, sampleErr := handler.AssignNextSample(dataset.ID, 0)
	is.NoErr(sampleErr)
	is.Equal(assignedSample.ID, dataset1Samples[3].ID)
}
