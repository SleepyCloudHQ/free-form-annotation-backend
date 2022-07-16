package models

import "gorm.io/gorm"

type UserRole string

const (
	AdminRole     UserRole = "admin"
	AnnotatorRole UserRole = "annotator"
)

type User struct {
	gorm.Model
	Email    string    `gorm:"unique" json:"email"`
	Password string    `json:"-"`
	Role     UserRole  `gorm:"default:annotator" json:"role"`
	Datasets []Dataset `gorm:"many2many:user_datasets" json:"-"`
}
