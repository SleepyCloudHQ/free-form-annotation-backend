package models

import (
	"errors"
	"gopkg.in/guregu/null.v4"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type StatusType string

const (
	Accepted  StatusType = "accepted"
	Rejected  StatusType = "rejected"
	Uncertain StatusType = "uncertain"
)

func (st StatusType) IsValid() error {
	switch st {
	case Accepted, Rejected, Uncertain:
		return nil
	}
	return errors.New("invalid status type")
}

func (st StatusType) ToNullString() null.String {
	return null.StringFrom(string(st))
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
	DatasetID   uint           `json:"dataset_id"`
	Annotations datatypes.JSON `json:"annotations"`
	Metadata    datatypes.JSON `json:"metadata"`
	Status      null.String    `json:"status"`
	Text        string         `json:"text"`
	AssignedTo  null.Int
}
