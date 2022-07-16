package handlers

import (
	"backend/app/models"
	"errors"
	"gopkg.in/guregu/null.v4"
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
	updateData := models.Sample{Annotations: patchRequest.Annotations, Metadata: patchRequest.Metadata, Status: patchRequest.Status.ToNullString()}
	if dbErr := s.DB.Where("dataset_id = ?", datasetId).First(&sample, sampleId).Updates(updateData).Error; dbErr != nil {
		return nil, dbErr
	}

	return sample, nil
}

func (s *SampleHandler) findUnassignedSample(datasetId uint) (*models.Sample, error) {
	sample := &models.Sample{}
	if dbErr := s.DB.Where("status IS NULL AND dataset_id = ? AND assigned_to IS NULL", datasetId).First(&sample).Error; dbErr != nil {
		return nil, dbErr
	}

	return sample, nil
}

func (s *SampleHandler) findAssignedSample(datasetId uint, userId uint) (*models.Sample, error) {
	sample := &models.Sample{}
	if dbErr := s.DB.Where("status IS NULL AND dataset_id = ? AND assigned_to = ?", datasetId, userId).First(&sample).Error; dbErr != nil {
		return nil, dbErr
	}

	return sample, nil
}

func (s *SampleHandler) findAndAssignSample(datasetId uint, userId uint) (*models.Sample, error) {
	// find unassigned sample
	unassignedSample, err := s.findUnassignedSample(datasetId)
	if err != nil {
		return nil, err
	}

	// assign the sample to the user
	if updateErr := s.DB.Model(&unassignedSample).Updates(models.Sample{AssignedTo: null.IntFrom(int64(userId))}).Error; updateErr != nil {
		return nil, updateErr
	}
	return unassignedSample, nil
}

func (s *SampleHandler) AssignNextSample(datasetId uint, userId uint) (*models.Sample, error) {
	// find already assigned sample
	assignedSample, assignedSampleErr := s.findAssignedSample(datasetId, userId)
	if assignedSampleErr != nil {
		if errors.Is(assignedSampleErr, gorm.ErrRecordNotFound) {
			// if there isn't an already assigned sample then assign a new one
			return s.findAndAssignSample(datasetId, userId)
		} else {
			return nil, assignedSampleErr
		}
	}

	return assignedSample, nil
}