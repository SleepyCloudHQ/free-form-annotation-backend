package main

import (
	"backend/app/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
)

func main() {
	db, err := gorm.Open(sqlite.Open("../../test.db"))
	if err != nil {
		log.Fatal(err)
	}
	if datasetErr := db.AutoMigrate(&models.Dataset{}); datasetErr != nil {
		log.Fatal(datasetErr)
		return
	}

	if sampleErr := db.AutoMigrate(&models.Sample{}); sampleErr != nil {
		log.Fatal(sampleErr)
		return
	}
}
