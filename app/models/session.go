package models

import (
	"gorm.io/gorm"
)

type SessionType string

const (
	EntityAnnotation   SessionType = "entity"
	RelationAnnotation SessionType = "relation"
)

type Session struct {
	gorm.Model
	Name    string      `gorm:"not null;" json:"name"`
	Samples []Sample    `json:"samples"`
	Type    SessionType `gorm:"not null" json:"type"`
}
