package handlers

import (
	"backend/app/models"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type DatasetStats struct {
	TotalSamples     int64 `json:"total_samples"`
	CompletedSamples int64 `json:"completed_samples"`
	PendingSamples   int64 `json:"pending_samples"`
}

type DatasetData struct {
	ID        uint
	Name      string             `json:"name"`
	Type      models.DatasetType `json:"type"`
	CreatedAt time.Time          `json:"created_at"`
	Metadata  datatypes.JSON     `json:"metadata"`
	Stats     *DatasetStats      `json:"stats"`
}

type DatasetsHandler struct {
	DB *gorm.DB
}

func NewDatasetsHandler(db *gorm.DB) *DatasetsHandler {
	return &DatasetsHandler{
		DB: db,
	}
}

func (s *DatasetsHandler) GetDatasets() []*DatasetData {
	var datasets []*models.Dataset
	s.DB.Order("created_at desc").Find(&datasets)

	result := make([]*DatasetData, len(datasets))
	for i, dataset := range datasets {
		result[i] = s.mapDatasetToDatasetData(dataset)
	}
	return result
}

func (s *DatasetsHandler) GetDatasetsForUser(user *models.User) ([]*DatasetData, error) {
	var datasets []*models.Dataset
	if dbErr := s.DB.Model(user).Association("Datasets").Find(&datasets); dbErr != nil {
		return nil, dbErr
	}

	result := make([]*DatasetData, len(datasets))
	for i, dataset := range datasets {
		result[i] = s.mapDatasetToDatasetData(dataset)
	}
	return result, nil

}

func (s *DatasetsHandler) GetDataset(id uint) (*DatasetData, error) {
	dataset := &models.Dataset{}
	if dbErr := s.DB.First(dataset, id).Error; dbErr != nil {
		return nil, dbErr
	}

	return s.mapDatasetToDatasetData(dataset), nil
}

func (s *DatasetsHandler) mapDatasetToDatasetData(dataset *models.Dataset) *DatasetData {
	return &DatasetData{
		ID:        dataset.ID,
		Name:      dataset.Name,
		Type:      dataset.Type,
		CreatedAt: dataset.CreatedAt,
		Metadata:  dataset.Metadata,
		Stats:     s.getDatasetsStats(dataset),
	}
}

func (s *DatasetsHandler) getDatasetsStats(dataset *models.Dataset) *DatasetStats {
	stats := &DatasetStats{
		TotalSamples:     0,
		CompletedSamples: 0,
		PendingSamples:   0,
	}

	completed := []models.StatusType{models.Accepted, models.Rejected, models.Uncertain}

	stats.TotalSamples = s.DB.Model(dataset).Association("Samples").Count()
	stats.CompletedSamples = s.DB.Model(dataset).Where("status IN ?", completed).Association("Samples").Count()
	stats.PendingSamples = s.DB.Model(dataset).Where("status IS NULL AND assigned_to IS NULL").Association("Samples").Count()

	return stats
}

func (s *DatasetsHandler) DeleteDataset(id uint) error {
	return s.DB.Delete(&models.Dataset{}, id).Error
}
