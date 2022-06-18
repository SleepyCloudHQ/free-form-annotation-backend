package main

import (
	"backend/app/models"
	"fmt"
	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
)

func main() {
	db, err := gorm.Open(sqlite.Open("../../test.db"))
	if err != nil {
		log.Fatal(err)
	}

	samples := []models.Sample{
		{
			Annotations: nil,
			Status:      models.Rejected,
			Data:        "This is a rejected sample",
		},
		{
			Annotations: datatypes.JSON([]byte(`{"some_key": "some_value"}`)),
			Status:      models.Accepted,
			Data:        "This is an accepted sample",
		},
		{
			Annotations: datatypes.JSON([]byte(`{"some_key": "some_value"}`)),
			Status:      models.Uncertain,
			Data:        "This is an uncertain sample",
		},
		{
			Annotations: nil,
			Status:      models.Unvisited,
			Data:        "This is a new sample",
		},
	}

	createErr := db.Create(&models.Session{
		Name:    "Session name",
		Type:    models.EntityAnnotation,
		Samples: samples,
	}).Error

	if createErr != nil {
		log.Fatal(createErr)
		return
	}

	fmt.Println("Test data created!")
}
