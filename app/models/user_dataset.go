package models

type UserDataset struct {
	UserID    int `gorm:"primaryKey"`
	DatasetID int `gorm:"primaryKey"`
}
