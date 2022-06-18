package models

import (
	"errors"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type StatusType string

const (
	Accepted  StatusType = "accepted"
	Rejected  StatusType = "rejected"
	Uncertain StatusType = "uncertain"
	Unvisited StatusType = "unvisited"
	Assigned  StatusType = "assigned"
)

func (st StatusType) IsValid() error {
	switch st {
	case Accepted, Rejected, Uncertain, Unvisited, Assigned:
		return nil
	}
	return errors.New("invalid status type")
}

//func (r *StatusType) Scan(value interface{}) error {
//	*r = StatusType(value.([]byte))
//	return nil
//}
//
//func (r StatusType) Value() (driver.Value, error) {
//	return string(r), nil
//}

type Sample struct {
	gorm.Model
	SessionID   uint
	Annotations datatypes.JSON
	Status      StatusType `gorm:"default:unvisited" json:"status"`
	Data        string
}
