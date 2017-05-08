package model

type CombinedModel struct {
	UUID     string       `json:"uuid"`
	Content  ContentModel `json:"content"`
	Metadata []Annotation `json:"metadata"`
}
