package dataset_export

import (
	"backend/app/models"
	"encoding/json"

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

type BoxPosition struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type Relationship struct {
	Id          uint        `json:"id"`
	Entity1     uint        `json:"entity1"`
	Entity2     uint        `json:"entity2"`
	Name        string      `json:"name"`
	Color       null.String `json:"color"`
	BoxPosition BoxPosition `json:"boxPosition"`
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

func MapSampleToSampleData(sample *models.Sample) (*SampleData, error) {
	var annotations AnnotationData
	if parsingErr := json.Unmarshal(sample.Annotations, &annotations); parsingErr != nil {
		return nil, parsingErr
	}

	return &SampleData{
		Text:        sample.Text,
		Annotations: annotations,
		Status:      sample.Status,
	}, nil
}

func MapSamplesToSampleData(samples []*models.Sample) ([]*SampleData, error) {
	result := make([]*SampleData, len(samples))
	for i, s := range samples {
		sampleData, exportErr := MapSampleToSampleData(s)
		if exportErr != nil {
			return nil, exportErr
		}
		result[i] = sampleData
	}

	return result, nil
}
