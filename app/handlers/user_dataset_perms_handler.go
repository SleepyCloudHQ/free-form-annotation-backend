package handlers

import (
	"backend/app/models"
	"errors"

	"gorm.io/gorm"
)

type DatasetToUserPermsRequest struct {
	DatasetId uint `json:"dataset_id" validate:"required"`
}

type UserDatasetPermsHandler struct {
	db *gorm.DB
}

func NewUserDatasetPermsHandler(db *gorm.DB) *UserDatasetPermsHandler {
	return &UserDatasetPermsHandler{
		db: db,
	}
}

func (u *UserDatasetPermsHandler) AddDatasetToUserPerms(userId uint, datasetId uint) error {
	userDatasetPerm := &models.UserDataset{UserID: userId, DatasetID: datasetId}
	return u.db.Create(userDatasetPerm).Error
}

func (u *UserDatasetPermsHandler) DeleteDatasetToUserPerms(userId uint, datasetId uint) error {
	userDatasetPerm := &models.UserDataset{UserID: userId, DatasetID: datasetId}
	return u.db.Delete(userDatasetPerm).Error
}

func (u *UserDatasetPermsHandler) UserHasDatasetPerms(userId uint, datasetId uint) (bool, error) {
	// check if the user has been assigned to the given dataset
	userDataset := &models.UserDataset{
		UserID:    userId,
		DatasetID: datasetId,
	}

	if result := u.db.First(userDataset); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, result.Error
	}

	return true, nil
}
