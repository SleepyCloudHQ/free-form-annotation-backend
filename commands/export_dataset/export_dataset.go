package main

import (
	"backend/app/handlers"
	"backend/app/models"
	"backend/app/utils"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gopkg.in/guregu/null.v4"
)

type Entity struct {
	Id      uint        `json:"id"`
	Content string      `json:"content"`
	Start   uint        `json:"start"`
	End     uint        `json:"end"`
	Tag     null.String `json:"tag"`
	Notes   null.String `json:"notes"`
	Color   null.String `json:"color"`
}

type Relationship struct {
	Id      uint        `json:"id"`
	Entity1 uint        `json:"entity1"`
	Entity2 uint        `json:"entity2"`
	Name    string      `json:"name"`
	Color   null.String `json:"color"`
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

func mapSampleToSampleData(sample *models.Sample) *SampleData {
	var annotations AnnotationData
	if parsingErr := json.Unmarshal(sample.Annotations, &annotations); parsingErr != nil {
		log.Fatal(parsingErr)
	}

	return &SampleData{
		Text:        sample.Text,
		Annotations: annotations,
		Status:      sample.Status,
	}
}

func main() {
	godotenv.Load()

	datasetId := flag.Int("d", -1, "dataset's id")
	outputFilePath := flag.String("f", "", "output file path")
	flag.Parse()

	if *datasetId < 0 {
		log.Fatal("Invalid dataset id")
	}

	if *outputFilePath == "" {
		log.Fatal("Invalid output file")
	}

	db, err := utils.Init_db()
	if err != nil {
		log.Fatal(err)
	}

	if sqlDB, sqlErr := db.DB(); sqlErr == nil {
		defer sqlDB.Close()
	}

	sampleHandler := handlers.NewSamplesHandler(db)
	samples, samplesErr := sampleHandler.GetSamples(uint(*datasetId))
	if samplesErr != nil {
		log.Fatal(samplesErr)
	}

	result := make([]*SampleData, len(*samples))
	for i, s := range *samples {
		result[i] = mapSampleToSampleData(&s)
	}

	outputFile, openFileErr := os.Create(*outputFilePath)
	if openFileErr != nil {
		log.Fatal(openFileErr)
	}

	defer outputFile.Close()

	if e := json.NewEncoder(outputFile).Encode(result); e != nil {
		log.Fatal(e)
	}

	fmt.Println("Dataset exported!")
}
