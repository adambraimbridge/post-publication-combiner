package model

// check if this is right to be a type... shouldn't i ember it in a struct?
// original: type annotations []annotation
type Annotations struct {
	Annotations []Annotation `json:"suggestions"`
	UUID        string       `json:"uuid"`
}

//Annotation is the main struct used to create and return structures
type Annotation struct {
	Thing `json:"thing,omitempty"`
}

//Thing represents a concept being linked to
type Thing struct {
	ID        string   `json:"id,omitempty"`
	PrefLabel string   `json:"prefLabel,omitempty"`
	Types     []string `json:"types,omitempty"`
	Predicate string   `json:"predicate,omitempty"`

	//INFO from the public-annotations-api
	ApiUrl          string   `json:"apiUrl,omitempty"`
	LeiCode         string   `json:"leiCode,omitempty"`
	FactsetID       string   `json:"factsetID,omitempty"`
	TmeIDs          []string `json:"tmeIDs,omitempty"`
	UUIDs           []string `json:"uuids,omitempty"`
	PlatformVersion string   `json:"platformVersion,omitempty"`
}
