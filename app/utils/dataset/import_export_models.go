package dataset

import (
	"gopkg.in/guregu/null.v4"
)

type Tag struct {
	Name  string      `json:"name"`
	Color null.String `json:"color"`
}

type Metadata struct {
	EntityTags       []Tag `json:"entityTags"`
	RelationshipTags []Tag `json:"relationshipTags"`
}

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
	Status      null.String    `json:"status"`
	Metadata    Metadata       `json:"metadata"`
}

type JsonDataset struct {
	Name     string       `json:"name"`
	Samples  []SampleData `json:"samples"`
	Metadata Metadata     `json:"metadata"`
}
