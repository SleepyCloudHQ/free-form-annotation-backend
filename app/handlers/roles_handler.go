package handlers

import "gorm.io/gorm"

type RolesHandler struct {
	DB *gorm.DB
}

func NewRolesHandler(db *gorm.DB) *RolesHandler {
	return &RolesHandler{
		DB: db,
	}
}
