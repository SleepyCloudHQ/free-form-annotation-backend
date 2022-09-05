package dataset_export

import (
	"backend/app/models"
	dataset_utils "backend/app/utils/dataset"
	"encoding/json"
)

func MapSampleToSampleData(sample *models.Sample) (*dataset_utils.SampleData, error) {
	var annotations dataset_utils.AnnotationData
	if parsingErr := json.Unmarshal(sample.Annotations, &annotations); parsingErr != nil {
		return nil, parsingErr
	}

	var metadata dataset_utils.Metadata
	if parsingErr := json.Unmarshal(sample.Metadata, &metadata); parsingErr != nil {
		return nil, parsingErr
	}

	return &dataset_utils.SampleData{
		Text:        sample.Text,
		Annotations: annotations,
		Status:      sample.Status,
		Metadata:    metadata,
	}, nil
}

func MapSamplesToSampleData(samples []*models.Sample) ([]dataset_utils.SampleData, error) {
	result := make([]dataset_utils.SampleData, len(samples))
	for i, s := range samples {
		sampleData, exportErr := MapSampleToSampleData(s)
		if exportErr != nil {
			return nil, exportErr
		}
		result[i] = *sampleData
	}

	return result, nil
}

func ExportDataset(dataset *models.Dataset, samples []*models.Sample) (*dataset_utils.JsonDataset, error) {
	var metadata dataset_utils.Metadata
	if parsingErr := json.Unmarshal(dataset.Metadata, &metadata); parsingErr != nil {
		return nil, parsingErr
	}

	sampleData, sampleDataErr := MapSamplesToSampleData(samples)
	if sampleDataErr != nil {
		return nil, sampleDataErr
	}

	return &dataset_utils.JsonDataset{
		Name:     dataset.Name,
		Samples:  sampleData,
		Metadata: metadata,
	}, nil
}
