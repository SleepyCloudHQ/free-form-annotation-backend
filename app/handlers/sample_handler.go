package handlers

import (
	"backend/app/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type PatchSampleRequest struct {
	Status      models.StatusType `json:"status"`
	Annotations datatypes.JSON    `json:"annotations"`
}

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
	sample := &models.Sample{}

	if dbErr := s.DB.Where("session_id = ?", sessionId).First(&sample, sampleId).Error; dbErr != nil {
		return nil, dbErr
	}

	return sample, nil
}

func (s *SampleHandler) PatchSample(sessionId uint, sampleId uint, patchRequest *PatchSampleRequest) (*models.Sample, error) {
	sample := &models.Sample{}
	updateData := models.Sample{Annotations: patchRequest.Annotations, Status: patchRequest.Status}
	if dbErr := s.DB.Where("session_id = ?", sessionId).First(&sample, sampleId).Updates(updateData).Error; dbErr != nil {
		return nil, dbErr
	}

	return sample, nil
}

func (s *SampleHandler) AssignNextSample(sessionId uint) (*models.Sample, error) {
	sample := &models.Sample{}

	if dbErr := s.DB.Where("status = ? AND session_id = ?", models.Unvisited, sessionId).First(&sample).Error; dbErr != nil {
		return nil, dbErr
	}

	// update status
	if updateErr := s.DB.Model(&sample).Updates(models.Sample{Status: models.Assigned}).Error; updateErr != nil {
		return nil, updateErr
	}
	return sample, nil
}
