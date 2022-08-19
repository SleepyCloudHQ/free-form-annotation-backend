package handlers

import (
	"backend/app/models"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type PatchUserRoleRequest struct {
	Role models.UserRole `json:"role" validate:"required"`
}

type UsersHandler struct {
	DB        *gorm.DB
	Validator *validator.Validate
}

func NewUsersHandler(db *gorm.DB, validator *validator.Validate) *UsersHandler {
	return &UsersHandler{
		DB:        db,
		Validator: validator,
	}
}

func (u *UsersHandler) GetUsers() []*models.User {
	var users []*models.User
	u.DB.Find(&users)
	return users
}

func (u *UsersHandler) GetUsersWithDatasets() []*models.User {
	var users []*models.User
	u.DB.Preload("Datasets").Find(&users)
	return users
}

func (u *UsersHandler) PatchUserRole(userId uint, request *PatchUserRoleRequest) (*models.User, error) {
	if valErr := u.Validator.Struct(request); valErr != nil {
		return nil, valErr.(validator.ValidationErrors)
	}

	if valErr := request.Role.IsValid(); valErr != nil {
		return nil, valErr
	}

	user := &models.User{}
	if err := u.DB.First(user, userId).Update("role", request.Role).Error; err != nil {
		return nil, err
	}

	return user, nil
}
