package handlers

import (
	"backend/app/models"
	"gorm.io/gorm"
	"time"
)

type DatasetStats struct {
	TotalSamples     int64 `json:"total_samples"`
	CompletedSamples int64 `json:"completed_samples"`
	PendingSamples   int64 `json:"pending_samples"`
}

type DatasetData struct {
	ID        uint               `json:"id"`
	Name      string             `json:"name"`
	Type      models.DatasetType `json:"type"`
	CreatedAt time.Time          `json:"created_at"`
	Stats     *DatasetStats      `json:"stats"`
}

type DatasetHandler struct {
	DB *gorm.DB
}

func NewDatasetHandler(db *gorm.DB) *DatasetHandler {
	return &DatasetHandler{
		DB: db,
	}
}

func (s *DatasetHandler) GetDatasets() *[]DatasetData {
	var datasets []models.Dataset
	s.DB.Find(&datasets)

	result := make([]DatasetData, len(datasets))
	for i, dataset := range datasets {
		result[i] = *s.mapDatasetToDatasetData(&dataset)
	}
	return &result
}

func (s *DatasetHandler) GetDataset(id uint) (*DatasetData, error) {
	dataset := &models.Dataset{}
	if dbErr := s.DB.First(dataset, id).Error; dbErr != nil {
		return nil, dbErr
	}

	return s.mapDatasetToDatasetData(dataset), nil
}

func (s *DatasetHandler) mapDatasetToDatasetData(dataset *models.Dataset) *DatasetData {
	return &DatasetData{
		ID:        dataset.ID,
		Name:      dataset.Name,
		Type:      dataset.Type,
		CreatedAt: dataset.CreatedAt,
		Stats:     s.getDatasetsStats(dataset),
	}
}

func (s *DatasetHandler) getDatasetsStats(dataset *models.Dataset) *DatasetStats {
	stats := &DatasetStats{
		TotalSamples:     0,
		CompletedSamples: 0,
		PendingSamples:   0,
	}

	completed := []models.StatusType{models.Accepted, models.Rejected, models.Uncertain}

	stats.TotalSamples = s.DB.Model(dataset).Association("Samples").Count()
	stats.CompletedSamples = s.DB.Model(dataset).Where("status IN ?", completed).Association("Samples").Count()
	stats.PendingSamples = s.DB.Model(dataset).Where("status = ?", models.Unvisited).Association("Samples").Count()

	return stats
}
