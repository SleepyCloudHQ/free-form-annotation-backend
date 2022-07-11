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
			Text:        "This is a rejected sample",
		},
		{
			Annotations: datatypes.JSON([]byte(`{"some_key": "some_value"}`)),
			Status:      models.Accepted,
			Text:        "This is an accepted sample",
		},
		{
			Annotations: datatypes.JSON([]byte(`{"some_key": "some_value"}`)),
			Status:      models.Uncertain,
			Text:        "This is an uncertain sample",
		},
		{
			Annotations: datatypes.JSON([]byte(`{ "entities": [ { "id": 1, "start": 15, "end": 19, "tag": "ent1", "notes": null, "elementId": "entity-1", "content": "text", "mark": true }, { "id": 2, "start": 39, "end": 42, "tag": "ent1", "notes": null, "elementId": "entity-2", "content": "and", "mark": true }, { "id": 3, "start": 51, "end": 59, "tag": "ent1", "notes": null, "elementId": "entity-3", "content": "ks.This ", "mark": true } ], "relationships": [ { "id": 1, "entity1": 1, "entity2": 2, "name": "r3", "boxPosition": { "x": 340, "y": 64 } } ] }`)),
			Metadata:    datatypes.JSON([]byte(`{ "entityTags": [ "ent1", "ent2", "ent3", "asdf" ], "relationshipTags": [ "r1", "r2", "r3" ] }`)),
			Status:      models.Unvisited,
			Text:        "This is a long text with some entities and some links.This is a long text with some entities and some links. This is a long text with some entities and some links.",
		},
	}

	createErr := db.Create(&models.Dataset{
		Name:    "Dataset name",
		Type:    models.EntityAnnotation,
		Samples: samples,
	}).Error

	if createErr != nil {
		log.Fatal(createErr)
		return
	}

	fmt.Println("Test data created!")
}
