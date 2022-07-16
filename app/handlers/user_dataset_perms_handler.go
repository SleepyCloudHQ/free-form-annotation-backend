package handlers

import (
	"backend/app/models"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type DatasetToUserPermsRequest struct {
	DatasetId uint `json:"dataset_id" validate:"required"`
}

type UserDatasetPermsHandler struct {
	DB        *gorm.DB
	Validator *validator.Validate
}

func NewUserDatasetPermsHandler(db *gorm.DB, validator *validator.Validate) *UserDatasetPermsHandler {
	return &UserDatasetPermsHandler{
		DB:        db,
		Validator: validator,
	}
}

func (u *UserDatasetPermsHandler) AddDatasetToUserPerms(userId uint, request *DatasetToUserPermsRequest) error {
	if valErr := u.Validator.Struct(request); valErr != nil {
		return valErr.(validator.ValidationErrors)
	}

	userDatasetPerm := &models.UserDataset{UserID: userId, DatasetID: request.DatasetId}
	return u.DB.Create(userDatasetPerm).Error
}

func (u *UserDatasetPermsHandler) DeleteDatasetToUserPerms(userId uint, request *DatasetToUserPermsRequest) error {
	if valErr := u.Validator.Struct(request); valErr != nil {
		return valErr.(validator.ValidationErrors)
	}

	userDatasetPerm := &models.UserDataset{UserID: userId, DatasetID: request.DatasetId}
	return u.DB.Delete(userDatasetPerm).Error
}
