package models

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type DatasetType string

const (
	EntityAnnotation   DatasetType = "entity"
	RelationAnnotation DatasetType = "relation"
)

type Dataset struct {
	gorm.Model
	Name     string         `gorm:"not null;" json:"name"`
	Samples  []Sample       `json:"samples"`
	Type     DatasetType    `gorm:"not null" json:"type"`
	Metadata datatypes.JSON `json:"metadata"`
}
