package dataset_import

import (
	"backend/app/models"
	"encoding/json"
	"io"
	"strings"

	"gopkg.in/guregu/null.v4"
	"gorm.io/datatypes"
)

type Entity struct {
	Id    uint        `json:"id"`
	Start uint        `json:"start"`
	End   uint        `json:"end"`
	Tag   null.String `json:"tag"`
	Notes null.String `json:"notes"`
	Color null.String `json:"color"`
}

type BoxPosition struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type Relationship struct {
	Id          uint         `json:"id"`
	Entity1     uint         `json:"entity1"`
	Entity2     uint         `json:"entity2"`
	Name        string       `json:"name"`
	Color       null.String  `json:"color"`
	BoxPosition *BoxPosition `json:"boxPosition"`
}

type AnnotationData struct {
	Entities      []Entity       `json:"entities"`
	Relationships []Relationship `json:"relationships"`
}

type SampleData struct {
	Text        string         `json:"text"`
	Annotations AnnotationData `json:"annotations"`
	// Status      null.String    `json:"status"`
}

func MapSampleDataToSample(sampleData []SampleData, datasetId uint) ([]models.Sample, error) {
	samples := make([]models.Sample, len(sampleData))

	for i, d := range sampleData {
		annotationsData, err := json.Marshal(d.Annotations)
		if err != nil {
			return nil, err
		}

		samples[i] = models.Sample{
			DatasetID:   datasetId,
			Annotations: datatypes.JSON(annotationsData),
			// Status:      d.Status,
			Text: d.Text,
		}
	}

	return samples, nil
}

func LoadSampleData(r io.Reader) ([]SampleData, error) {
	var samplesData []SampleData
	parsingErr := json.NewDecoder(r).Decode(&samplesData)
	if parsingErr != nil {
		return nil, parsingErr
	}

	return samplesData, nil
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

func ParseTags(inputData string) []string {
	if inputData == "" {
		return nil
	}

	return strings.Split(inputData, ",")
}
