package handlers

import (
	"backend/app/models"
	"gorm.io/gorm"
)

type DatasetToUserPermsRequest struct {
	DatasetId uint `json:"dataset_id" validate:"required"`
}

type UserDatasetPermsHandler struct {
	DB *gorm.DB
}

func NewUserDatasetPermsHandler(db *gorm.DB) *UserDatasetPermsHandler {
	return &UserDatasetPermsHandler{
		DB: db,
	}
}

func (u *UserDatasetPermsHandler) AddDatasetToUserPerms(userId uint, datasetId uint) error {
	userDatasetPerm := &models.UserDataset{UserID: userId, DatasetID: datasetId}
	return u.DB.Create(userDatasetPerm).Error
}

func (u *UserDatasetPermsHandler) DeleteDatasetToUserPerms(userId uint, datasetId uint) error {
	userDatasetPerm := &models.UserDataset{UserID: userId, DatasetID: datasetId}
	return u.DB.Delete(userDatasetPerm).Error
}
