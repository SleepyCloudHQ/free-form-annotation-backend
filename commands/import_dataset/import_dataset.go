package main

import (
	"backend/app/models"
	"backend/app/utils"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

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

func mapSampleDataToSample(sampleData []SampleData, datasetId uint) []models.Sample {
	samples := make([]models.Sample, len(sampleData))

	for i, d := range sampleData {
		annotationsData, err := json.Marshal(d.Annotations)
		if err != nil {
			log.Fatal(err)
		}

		samples[i] = models.Sample{
			DatasetID:   datasetId,
			Annotations: datatypes.JSON(annotationsData),
			Status:      d.Status,
			Text:        d.Text,
		}
	}

	return samples
}

func loadSampleData(filePath string) ([]SampleData, error) {
	samplesFile, fileErr := os.Open(filePath)
	if fileErr != nil {
		return nil, fileErr
	}

	var samplesData []SampleData
	parsingErr := json.NewDecoder(samplesFile).Decode(&samplesData)
	if parsingErr != nil {
		return nil, parsingErr
	}

	return samplesData, nil
}

func createDatasetMetadata(entityTags []string, relationshipTags []string) (datatypes.JSON, error) {
	metadata := struct {
		EntityTags       []string `json:"entityTags"`
		RelationshipTags []string `json:"relationshipTags"`
	}{
		EntityTags:       entityTags,
		RelationshipTags: relationshipTags,
	}
	metadataJson, marshalErr := json.Marshal(metadata)
	if marshalErr != nil {
		return nil, marshalErr
	}
	return datatypes.JSON([]byte(metadataJson)), nil
}

func parseTags(inputData string) []string {
	if inputData == "" {
		return nil
	}

	return strings.Split(inputData, ",")
}

func main() {
	datasetName := flag.String("n", "", "dataset's name")
	samplesFilePath := flag.String("f", "", "samples file path")

	predefinedEntities := flag.String("e", "", "entities")
	predefinedRelationships := flag.String("r", "", "relationships")

	flag.Parse()

	if *datasetName == "" {
		log.Fatal("Dataset name cannot be empty")
	}

	entityTags := parseTags(*predefinedEntities)
	relationshipTags := parseTags(*predefinedRelationships)
	metadata, metadataErr := createDatasetMetadata(entityTags, relationshipTags)
	if metadataErr != nil {
		log.Fatal(metadataErr)
	}

	samplesData, samplesDataErr := loadSampleData(*samplesFilePath)
	if samplesDataErr != nil {
		log.Fatal(samplesDataErr)
	}

	db, dbErr := utils.Init_db()
	if dbErr != nil {
		log.Fatal(dbErr)
	}

	// create dataset
	dataset := &models.Dataset{
		Name:     *datasetName,
		Metadata: metadata,
	}

	if datasetCreateErr := db.Create(&dataset).Error; datasetCreateErr != nil {
		log.Fatal(datasetCreateErr)
	}

	samples := mapSampleDataToSample(samplesData, dataset.ID)
	// create samples in a batch
	if sampleCreateErr := db.Create(&samples).Error; sampleCreateErr != nil {
		log.Fatal(sampleCreateErr)
	}

	fmt.Println("Dataset created!")
}
