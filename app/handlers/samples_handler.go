package handlers

import (
	"backend/app/models"
	"errors"
	"gopkg.in/guregu/null.v4"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type PatchSampleRequest struct {
	Status      null.String    `json:"status"`
	Annotations datatypes.JSON `json:"annotations"`
	Metadata    datatypes.JSON `json:"metadata"`
}

type SamplesHandler struct {
	DB *gorm.DB
}

func NewSamplesHandler(db *gorm.DB) *SamplesHandler {
	return &SamplesHandler{
		DB: db,
	}
}

func (s *SamplesHandler) GetSamples(datasetId uint) ([]*models.Sample, error) {
	var samples []*models.Sample
	if dbErr := s.DB.Where("dataset_id = ?", datasetId).Find(&samples).Error; dbErr != nil {
		return nil, dbErr
	}

	return samples, nil
}

func (s *SamplesHandler) GetSamplesWithStatus(datasetId uint, status models.StatusType) ([]*models.Sample, error) {
	var samples []*models.Sample
	if dbErr := s.DB.Where("dataset_id = ? AND status = ?", datasetId, status).Find(&samples).Error; dbErr != nil {
		return nil, dbErr
	}

	return samples, nil
}

func (s *SamplesHandler) GetSample(datasetId uint, sampleId uint) (*models.Sample, error) {
	sample := &models.Sample{}

	if dbErr := s.DB.Where("dataset_id = ?", datasetId).First(&sample, sampleId).Error; dbErr != nil {
		return nil, dbErr
	}

	return sample, nil
}

func (s *SamplesHandler) PatchSample(datasetId uint, sampleId uint, patchRequest *PatchSampleRequest) (*models.Sample, error) {
	sample := &models.Sample{}
	updateData := models.Sample{Annotations: patchRequest.Annotations, Metadata: patchRequest.Metadata, Status: patchRequest.Status}
	if dbErr := s.DB.Where("dataset_id = ?", datasetId).First(&sample, sampleId).Updates(updateData).Error; dbErr != nil {
		return nil, dbErr
	}

	return sample, nil
}

func (s *SamplesHandler) findUnassignedSample(datasetId uint) (*models.Sample, error) {
	sample := &models.Sample{}
	if dbErr := s.DB.Where("status IS NULL AND dataset_id = ? AND assigned_to IS NULL", datasetId).First(&sample).Error; dbErr != nil {
		return nil, dbErr
	}

	return sample, nil
}

func (s *SamplesHandler) findAssignedSample(datasetId uint, userId uint) (*models.Sample, error) {
	sample := &models.Sample{}
	if dbErr := s.DB.Where("status IS NULL AND dataset_id = ? AND assigned_to = ?", datasetId, userId).First(&sample).Error; dbErr != nil {
		return nil, dbErr
	}

	return sample, nil
}

func (s *SamplesHandler) findAndAssignSample(datasetId uint, userId uint) (*models.Sample, error) {
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

func (s *SamplesHandler) AssignNextSample(datasetId uint, userId uint) (*models.Sample, error) {
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
