package models

type UserDataset struct {
	UserID    uint `gorm:"primaryKey"`
	DatasetID uint `gorm:"primaryKey"`
}
