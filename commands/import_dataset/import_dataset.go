package main

import (
	"backend/app/models"
	"backend/app/utils"
	dataset_import "backend/app/utils/dataset/import"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func loadSampleDataFromFile(filePath string) ([]dataset_import.SampleData, error) {
	samplesFile, fileErr := os.Open(filePath)
	if fileErr != nil {
		return nil, fileErr
	}
	return dataset_import.LoadSampleData(samplesFile)
}

func main() {
	godotenv.Load()
	datasetName := flag.String("n", "", "dataset's name")
	samplesFilePath := flag.String("f", "", "samples file path")

	predefinedEntities := flag.String("e", "", "entities")
	predefinedRelationships := flag.String("r", "", "relationships")

	flag.Parse()

	if *datasetName == "" {
		log.Fatal("Dataset name cannot be empty")
	}

	entityTags := dataset_import.ParseTags(*predefinedEntities)
	relationshipTags := dataset_import.ParseTags(*predefinedRelationships)
	metadata, metadataErr := dataset_import.CreateDatasetMetadata(entityTags, relationshipTags)
	if metadataErr != nil {
		log.Fatal(metadataErr)
	}

	samplesData, samplesDataErr := loadSampleDataFromFile(*samplesFilePath)
	if samplesDataErr != nil {
		log.Fatal(samplesDataErr)
	}

	db, dbErr := utils.Init_db()
	if dbErr != nil {
		log.Fatal(dbErr)
	}

	if sqlDB, sqlErr := db.DB(); sqlErr == nil {
		defer sqlDB.Close()
	}

	// create dataset
	dataset := &models.Dataset{
		Name:     *datasetName,
		Metadata: metadata,
	}

	if datasetCreateErr := db.Create(&dataset).Error; datasetCreateErr != nil {
		log.Fatal(datasetCreateErr)
	}

	samples, samplesErr := dataset_import.MapSampleDataToSample(samplesData, dataset.ID)
	if samplesErr != nil {
		log.Fatal(samplesErr)
	}

	// create samples in a batch
	if sampleCreateErr := db.Create(&samples).Error; sampleCreateErr != nil {
		log.Fatal(sampleCreateErr)
	}

	fmt.Printf("Dataset (ID: %d) created!\n", dataset.ID)
}
