package main

import (
	"backend/app/models"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"gopkg.in/guregu/null.v4"
	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Entity struct {
	Id    uint        `json:"id"`
	Start uint        `json:"start"`
	End   uint        `json:"end"`
	Tag   null.String `json:"tag"`
	Notes null.String `json:"notes"`
}

type Relationship struct {
	Id      uint   `json:"id"`
	Entity1 uint   `json:"entity1"`
	Entity2 uint   `json:"entity2"`
	Name    string `json:"name"`
}

type AnnotationData struct {
	Entities      []Entity       `json:"entities"`
	Relationships []Relationship `json:"relationships"`
}

type SampleData struct {
	Text        string         `json:"text"`
	Annotations AnnotationData `json:"annotations"`
	Status      null.String    `json:"status"`
}

func mapSampleDataToSample(sampleData *SampleData, datasetId uint) *models.Sample {
	annotationsData, err := json.Marshal(sampleData.Annotations)
	if err != nil {
		log.Fatal(err)
	}

	return &models.Sample{
		DatasetID:   datasetId,
		Annotations: datatypes.JSON(annotationsData),
		Status:      sampleData.Status,
		Text:        sampleData.Text,
	}
}

func main() {
	datasetName := flag.String("n", "", "dataset's name")
	samplesFilePath := flag.String("f", "", "samples file path")
	flag.Parse()

	if *datasetName == "" {
		log.Fatal("Dataset name cannot be empty")
	}

	samplesFile, fileErr := os.Open(*samplesFilePath)
	if fileErr != nil {
		log.Fatal(fileErr)
	}

	var samplesData []SampleData
	parsingErr := json.NewDecoder(samplesFile).Decode(&samplesData)
	if parsingErr != nil {
		log.Fatal(parsingErr)
	}

	db, dbErr := gorm.Open(sqlite.Open("../../test.db"))
	if dbErr != nil {
		log.Fatal(dbErr)
	}

	// create dataset
	dataset := &models.Dataset{
		Name: *datasetName,
	}

	if datasetCreateErr := db.Create(&dataset).Error; datasetCreateErr != nil {
		log.Fatal(datasetCreateErr)
	}

	samples := make([]*models.Sample, len(samplesData))
	for i, d := range samplesData {
		samples[i] = mapSampleDataToSample(&d, dataset.ID)
	}

	// create samples in a batch
	if sampleCreateErr := db.Create(&samples).Error; sampleCreateErr != nil {
		log.Fatal(sampleCreateErr)
	}

	fmt.Println("Dataset created!")
}
