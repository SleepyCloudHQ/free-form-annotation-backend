package main

import (
	"backend/app/handlers"
	"backend/app/utils"
	dataset_export "backend/app/utils/dataset/export"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

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

	samplesData, exportErr := dataset_export.MapSamplesToSampleData(samples)
	if exportErr != nil {
		log.Fatal(exportErr)
	}

	outputFile, openFileErr := os.Create(*outputFilePath)
	if openFileErr != nil {
		log.Fatal(openFileErr)
	}

	defer outputFile.Close()

	if e := json.NewEncoder(outputFile).Encode(samplesData); e != nil {
		log.Fatal(e)
	}

	fmt.Println("Dataset exported!")
}
