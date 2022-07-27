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

	"gopkg.in/guregu/null.v4"
	"gorm.io/datatypes"
)

type SampleData struct {
	Text        string         `json:"text"`
	Annotations datatypes.JSON `json:"annotations"`
	Status      null.String    `json:"status"`
}

func mapSampleToSampleData(sample *models.Sample) *SampleData {
	return &SampleData{
		Text:        sample.Text,
		Annotations: sample.Annotations,
		Status:      sample.Status,
	}
}

func main() {
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

	sampleHandler := handlers.NewSamplesHandler(db)
	samples, samplesErr := sampleHandler.GetSamples(uint(*datasetId))
	if samplesErr != nil {
		log.Fatal(samplesErr)
	}

	result := make([]*SampleData, len(*samples))
	for i, s := range *samples {
		result[i] = mapSampleToSampleData(&s)
	}

	outputFile, openFileErr := os.OpenFile(*outputFilePath, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if openFileErr != nil {
		log.Fatal(openFileErr)
	}

	defer outputFile.Close()

	if e := json.NewEncoder(outputFile).Encode(result); e != nil {
		log.Fatal(e)
	}

	fmt.Println("Dataset exported!")
}
