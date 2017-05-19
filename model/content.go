package model

type MessageContent struct {
	ContentURI   string       `json:"contentUri"`
	ContentModel ContentModel `json:"payload"`
	LastModified string       `json:"lastModified"`
}

type ContentModel struct {
	UUID               string       `json:"uuid"`
	Title              string       `json:"title"`
	Body               string       `json:"body"`
	Identifiers        []Identifier `json:"identifiers"`
	PublishedDate      string       `json:"publishedDate"`
	LastModified       string       `json:"lastModified"`
	FirstPublishedDate string       `json:"firstPublishedDate"`

	MediaType        string `json:"mediaType"` //image/set, image/jpeg, video/next-video-editor
	MarkedDeleted    bool   `json:"marked_deleted"`
	Byline           string `json:"byline"` // wordpress
	Standfirst       string `json:"standfirst"`
	Description      string `json:"description"`
	MainImage        string `json:"mainImage"`
	PublishReference string `json:"publishReference"`
	Type             string `json:"type"`
}

type Identifier struct {
	Authority       string `json:"authority"`
	IdentifierValue string `json:"identifierValue"`
}
