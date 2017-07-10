package model

type Annotations struct {
	Annotations []Annotation `json:"annotations"`
	UUID        string       `json:"uuid"`
}

type Annotation struct {
	Thing `json:"thing,omitempty"`
}

type Thing struct {
	ID              string   `json:"id,omitempty"`
	PrefLabel       string   `json:"prefLabel,omitempty"`
	Types           []string `json:"types,omitempty"`
	Predicate       string   `json:"predicate,omitempty"`
	ApiUrl          string   `json:"apiUrl,omitempty"`
	LeiCode         string   `json:"leiCode,omitempty"`
	FactsetID       string   `json:"factsetID,omitempty"`
	TmeIDs          []string `json:"tmeIDs,omitempty"`
	UUIDs           []string `json:"uuids,omitempty"`
	PlatformVersion string   `json:"platformVersion,omitempty"`
}
