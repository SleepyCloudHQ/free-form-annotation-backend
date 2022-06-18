package handlers

import (
	"backend/app/models"
	"gorm.io/gorm"
)

type SampleHandler struct {
	DB *gorm.DB
}

func NewSampleHandler(db *gorm.DB) *SampleHandler {
	return &SampleHandler{
		DB: db,
	}
}

func (s *SampleHandler) GetSamples(sessionId uint) (*[]models.Sample, error) {
	var samples []models.Sample
	if dbErr := s.DB.Where("session_id = ?", sessionId).Find(&samples).Error; dbErr != nil {
		return nil, dbErr
	}

	return &samples, nil
}

func (s *SampleHandler) GetSamplesWithStatus(sessionId uint, status models.StatusType) (*[]models.Sample, error) {
	var samples []models.Sample
	if dbErr := s.DB.Where("session_id = ? AND status = ?", sessionId, status).Find(&samples).Error; dbErr != nil {
		return nil, dbErr
	}

	return &samples, nil
}

func (s *SampleHandler) GetSample(sessionId uint, sampleId uint) (*models.Sample, error) {
	sample := &models.Sample{
		Model: gorm.Model{
			ID: sampleId,
		},
		SessionID: sessionId,
	}

	if dbErr := s.DB.First(&sample).Error; dbErr != nil {
		return nil, dbErr
	}

	return sample, nil
}
