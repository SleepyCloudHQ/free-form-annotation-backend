package handlers

import (
	"backend/app/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type PatchSampleRequest struct {
	Status      models.StatusType `json:"status"`
	Annotations datatypes.JSON    `json:"annotations"`
	Metadata    datatypes.JSON    `json:"metadata"`
}

type SampleHandler struct {
	DB *gorm.DB
}

func NewSampleHandler(db *gorm.DB) *SampleHandler {
	return &SampleHandler{
		DB: db,
	}
}

func (s *SampleHandler) GetSamples(datasetId uint) (*[]models.Sample, error) {
	var samples []models.Sample
	if dbErr := s.DB.Where("dataset_id = ?", datasetId).Find(&samples).Error; dbErr != nil {
		return nil, dbErr
	}

	return &samples, nil
}

func (s *SampleHandler) GetSamplesWithStatus(datasetId uint, status models.StatusType) (*[]models.Sample, error) {
	var samples []models.Sample
	if dbErr := s.DB.Where("dataset_id = ? AND status = ?", datasetId, status).Find(&samples).Error; dbErr != nil {
		return nil, dbErr
	}

	return &samples, nil
}

func (s *SampleHandler) GetSample(datasetId uint, sampleId uint) (*models.Sample, error) {
	sample := &models.Sample{}

	if dbErr := s.DB.Where("dataset_id = ?", datasetId).First(&sample, sampleId).Error; dbErr != nil {
		return nil, dbErr
	}

	return sample, nil
}

func (s *SampleHandler) PatchSample(datasetId uint, sampleId uint, patchRequest *PatchSampleRequest) (*models.Sample, error) {
	sample := &models.Sample{}
	updateData := models.Sample{Annotations: patchRequest.Annotations, Metadata: patchRequest.Metadata, Status: patchRequest.Status}
	if dbErr := s.DB.Where("dataset_id = ?", datasetId).First(&sample, sampleId).Updates(updateData).Error; dbErr != nil {
		return nil, dbErr
	}

	return sample, nil
}

func (s *SampleHandler) AssignNextSample(datasetId uint) (*models.Sample, error) {
	sample := &models.Sample{}

	if dbErr := s.DB.Where("status = ? AND dataset_id = ?", models.Unvisited, datasetId).First(&sample).Error; dbErr != nil {
		return nil, dbErr
	}

	// update status
	if updateErr := s.DB.Model(&sample).Updates(models.Sample{Status: models.Assigned}).Error; updateErr != nil {
		return nil, updateErr
	}
	return sample, nil
}
