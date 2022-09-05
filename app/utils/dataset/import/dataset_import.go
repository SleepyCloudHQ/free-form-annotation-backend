package dataset_import

import (
	"backend/app/models"
	"encoding/json"
	"io"

	dataset "backend/app/utils/dataset"
	"gorm.io/datatypes"
)

func MapSampleDataToSample(sampleData []dataset.SampleData, datasetId uint) ([]models.Sample, error) {
	samples := make([]models.Sample, len(sampleData))

	for i, d := range sampleData {
		annotationsData, err := json.Marshal(d.Annotations)
		if err != nil {
			return nil, err
		}

		metadata, metadataErr := json.Marshal(d.Metadata)
		if metadataErr != nil {
			return nil, err
		}

		samples[i] = models.Sample{
			DatasetID:   datasetId,
			Annotations: datatypes.JSON(annotationsData),
			Status:      d.Status,
			Text:        d.Text,
			Metadata:    metadata,
		}
	}

	return samples, nil
}

func CreateDatasetMetadata(entityTags []string, relationshipTags []string) (datatypes.JSON, error) {
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

func ParseDataset(r io.Reader) (*dataset.JsonDataset, error) {
	var dataset dataset.JsonDataset
	parsingErr := json.NewDecoder(r).Decode(&dataset)
	if parsingErr != nil {
		return nil, parsingErr
	}

	return &dataset, nil
}

func MarshalDatasetMetadata(metadata dataset.Metadata) (datatypes.JSON, error) {
	metadataJson, marshalErr := json.Marshal(metadata)
	if marshalErr != nil {
		return nil, marshalErr
	}
	return datatypes.JSON([]byte(metadataJson)), nil
}
