package model

type CombinedModel struct {
	UUID       string       `json:"uuid"`
	Content    ContentModel `json:"content"`
	V1Metadata []Annotation `json:"v1-metadata"`
}
