package handlers

import (
	"backend/app/models"
	"gorm.io/gorm"
)

type UsersHandler struct {
	DB *gorm.DB
}

func NewUsersHandler(db *gorm.DB) *UsersHandler {
	return &UsersHandler{
		DB: db,
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

func (u *UsersHandler) PatchUserRole(userId uint, role models.UserRole) (*models.User, error) {
	user := &models.User{}
	if err := u.DB.First(user, userId).Update("role", role).Error; err != nil {
		return nil, err
	}

	return user, nil
}
