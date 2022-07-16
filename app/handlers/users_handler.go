package handlers

import "gorm.io/gorm"

type UsersHandler struct {
	DB *gorm.DB
}

func NewUsersHandler(db *gorm.DB) *UsersHandler {
	return &UsersHandler{
		DB: db,
	}
}
