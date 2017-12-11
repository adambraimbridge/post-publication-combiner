package processor

import (
	"errors"
	"github.com/Financial-Times/post-publication-combiner/utils"
	"github.com/golang/go/src/pkg/fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

func TestGetCombinedModelForContent(t *testing.T) {

	tests := []struct {
		contentModel ContentModel
		retrievedAnn []Annotation
		retrievedErr error
		expModel     CombinedModel
		expError     error
	}{
		{
			ContentModel{},
			[]Annotation{},
			nil,
			CombinedModel{},
			errors.New("Content has no UUID provided. Can't deduce annotations for it."),
		},
		{
			ContentModel{
				UUID: "some uuid",
			},
			[]Annotation{},
			errors.New("some error"),
			CombinedModel{},
			errors.New("some error"),
		},
		{
			ContentModel{
				UUID: "some uuid",
			},
			[]Annotation{},
			errors.New("Could not unmarshall annotations for content with uuid"),
			CombinedModel{},
			errors.New("Could not unmarshall annotations for content with uuid"),
		},
		{
			ContentModel{
				UUID:  "622de808-3a7a-49bd-a7fb-2a33f64695be",
				Title: "Title",
				Body:  "<body>something relevant here</body>",
				Identifiers: []Identifier{
					{
						Authority:       "FTCOM-METHODE_identifier",
						IdentifierValue: "53217c65-ecef-426e-a3ac-3787e2e62e87",
					},
				},
				PublishedDate:      "2017-04-10T08:03:58.000Z",
				LastModified:       "2017-04-10T08:09:01.808Z",
				FirstPublishedDate: "2017-04-10T08:03:58.000Z",
				MediaType:          "mediaType",
				MarkedDeleted:      false,
				Byline:             "FT Reporters",
				Standfirst:         "A simple line with an article summary",
				Description:        "descr",
				MainImage:          "2934de46-5240-4c7d-8576-f12ae12e4a37",
				PublishReference:   "tid_unique_reference",
			},
			[]Annotation{
				{
					Thing: Thing{
						ID:        "http://base-url/80bec524-8c75-4d0f-92fa-abce3962d995",
						PrefLabel: "Barclays",
						Types: []string{"http://base-url/core/Thing",
							"http://base-url/concept/Concept",
							"http://base-url/organisation/Organisation",
							"http://base-url/company/Company",
							"http://base-url/company/PublicCompany",
						},
						Predicate: "http://base-url/about",
						ApiUrl:    "http://base-url/80bec524-8c75-4d0f-92fa-abce3962d995",
						LeiCode:   "leicode_id_1",
						FactsetID: "factset-id1",
						TmeIDs:    []string{"tme_id1"},
						UUIDs: []string{"80bec524-8c75-4d0f-92fa-abce3962d995",
							"factset-generated-uuid"},
						PlatformVersion: "v1",
					},
				},
			},
			nil,
			CombinedModel{
				UUID: "622de808-3a7a-49bd-a7fb-2a33f64695be",
				Content: ContentModel{
					UUID:  "622de808-3a7a-49bd-a7fb-2a33f64695be",
					Title: "Title",
					Body:  "<body>something relevant here</body>",
					Identifiers: []Identifier{
						{
							Authority:       "FTCOM-METHODE_identifier",
							IdentifierValue: "53217c65-ecef-426e-a3ac-3787e2e62e87",
						},
					},
					PublishedDate:      "2017-04-10T08:03:58.000Z",
					LastModified:       "2017-04-10T08:09:01.808Z",
					FirstPublishedDate: "2017-04-10T08:03:58.000Z",
					MediaType:          "mediaType",
					MarkedDeleted:      false,
					Byline:             "FT Reporters",
					Standfirst:         "A simple line with an article summary",
					Description:        "descr",
					MainImage:          "2934de46-5240-4c7d-8576-f12ae12e4a37",
					PublishReference:   "tid_unique_reference",
				},
				Metadata: []Annotation{
					{
						Thing: Thing{
							ID:        "http://base-url/80bec524-8c75-4d0f-92fa-abce3962d995",
							PrefLabel: "Barclays",
							Types: []string{"http://base-url/core/Thing",
								"http://base-url/concept/Concept",
								"http://base-url/organisation/Organisation",
								"http://base-url/company/Company",
								"http://base-url/company/PublicCompany",
							},
							Predicate: "http://base-url/about",
							ApiUrl:    "http://base-url/80bec524-8c75-4d0f-92fa-abce3962d995",
							LeiCode:   "leicode_id_1",
							FactsetID: "factset-id1",
							TmeIDs:    []string{"tme_id1"},
							UUIDs: []string{"80bec524-8c75-4d0f-92fa-abce3962d995",
								"factset-generated-uuid"},
							PlatformVersion: "v1",
						},
					},
				},
			},
			nil,
		},
	}

	for _, testCase := range tests {
		combiner := DataCombiner{
			MetadataRetriever: DummyMetadataRetriever{testCase.retrievedAnn, testCase.retrievedErr},
		}
		m, err := combiner.GetCombinedModelForContent(testCase.contentModel, "some_platform_version")
		assert.Equal(t, testCase.expModel, m,
			fmt.Sprintf("Expected model: %v was not equal with the received one: %v \n", testCase.expModel, m))
		if testCase.expError == nil {
			assert.Equal(t, nil, err)
		} else {
			assert.Contains(t, err.Error(), testCase.expError.Error())
		}
	}
}

func TestGetCombinedModelForAnnotations(t *testing.T) {

	tests := []struct {
		metadata            Annotations
		retrievedContent    ContentModel
		retreivedContentErr error
		retrievedAnn        []Annotation
		retreivedAnnErr     error
		expModel            CombinedModel
		expError            error
	}{
		{
			Annotations{},
			ContentModel{},
			nil,
			[]Annotation{},
			nil,
			CombinedModel{},
			errors.New("Annotations have no UUID referenced. Can't deduce content for it."),
		},
		{
			Annotations{UUID: "some_uuid"},
			ContentModel{},
			errors.New("some content error"),
			[]Annotation{},
			nil,
			CombinedModel{},
			errors.New("some content error"),
		},
		{
			Annotations{UUID: "some_uuid"},
			ContentModel{},
			errors.New("some content error"),
			[]Annotation{},
			errors.New("some metadata error"),
			CombinedModel{},
			errors.New("some content error"),
		},
		{
			Annotations{UUID: "some_uuid"},
			ContentModel{},
			nil,
			[]Annotation{},
			errors.New("some metadata error"),
			CombinedModel{},
			errors.New("some metadata error"),
		},
		{
			Annotations{UUID: "some_uuid"},
			ContentModel{
				UUID:  "some_uuid",
				Title: "title",
				Body:  "body",
			},
			nil,
			[]Annotation{
				{Thing{
					ID:        "http://base-url/80bec524-8c75-4d0f-92fa-abce3962d995",
					PrefLabel: "Barclays",
					Types: []string{"http://base-url/core/Thing",
						"http://base-url/concept/Concept",
					},
					Predicate: "http://base-url/about",
					ApiUrl:    "http://base-url/80bec524-8c75-4d0f-92fa-abce3962d995",
					TmeIDs:    []string{"tme_id1"},
					UUIDs: []string{"80bec524-8c75-4d0f-92fa-abce3962d995",
						"factset-generated-uuid"},
					PlatformVersion: "v1",
				},
				},
			},
			nil,
			CombinedModel{
				UUID: "some_uuid",
				Content: ContentModel{
					UUID:  "some_uuid",
					Title: "title",
					Body:  "body",
				},
				Metadata: []Annotation{
					{Thing{
						ID:        "http://base-url/80bec524-8c75-4d0f-92fa-abce3962d995",
						PrefLabel: "Barclays",
						Types: []string{"http://base-url/core/Thing",
							"http://base-url/concept/Concept",
						},
						Predicate: "http://base-url/about",
						ApiUrl:    "http://base-url/80bec524-8c75-4d0f-92fa-abce3962d995",
						TmeIDs:    []string{"tme_id1"},
						UUIDs: []string{"80bec524-8c75-4d0f-92fa-abce3962d995",
							"factset-generated-uuid"},
						PlatformVersion: "v1",
					},
					},
				},
			},
			nil,
		},
		{
			Annotations{UUID: "some_uuid"},
			ContentModel{
				UUID:  "some_uuid",
				Title: "title",
				Body:  "body",
				Type:  "Video",
				Identifiers: []Identifier{
					{
						Authority:       "http://api.ft.com/system/NEXT-VIDEO-EDITOR",
						IdentifierValue: "some_uuid",
					},
				},
			},
			nil,
			[]Annotation{
				{Thing{
					ID:        "http://base-url/80bec524-8c75-4d0f-92fa-abce3962d995",
					PrefLabel: "Barclays",
					Types: []string{"http://base-url/core/Thing",
						"http://base-url/concept/Concept",
					},
					Predicate: "http://base-url/about",
					ApiUrl:    "http://base-url/80bec524-8c75-4d0f-92fa-abce3962d995",
					TmeIDs:    []string{"tme_id1"},
					UUIDs: []string{"80bec524-8c75-4d0f-92fa-abce3962d995",
						"factset-generated-uuid"},
					PlatformVersion: "next-video",
				},
				},
			},
			nil,
			CombinedModel{
				UUID: "some_uuid",
				Content: ContentModel{
					UUID:  "some_uuid",
					Title: "title",
					Body:  "body",
					Type:  "Video",
					Identifiers: []Identifier{
						{
							Authority:       "http://api.ft.com/system/NEXT-VIDEO-EDITOR",
							IdentifierValue: "some_uuid",
						},
					},
				},
				Metadata: []Annotation{
					{Thing{
						ID:        "http://base-url/80bec524-8c75-4d0f-92fa-abce3962d995",
						PrefLabel: "Barclays",
						Types: []string{"http://base-url/core/Thing",
							"http://base-url/concept/Concept",
						},
						Predicate: "http://base-url/about",
						ApiUrl:    "http://base-url/80bec524-8c75-4d0f-92fa-abce3962d995",
						TmeIDs:    []string{"tme_id1"},
						UUIDs: []string{"80bec524-8c75-4d0f-92fa-abce3962d995",
							"factset-generated-uuid"},
						PlatformVersion: "next-video",
					},
					},
				},
			},
			nil,
		},
	}

	for _, testCase := range tests {
		combiner := DataCombiner{
			ContentRetriever:  DummyContentRetriever{testCase.retrievedContent, testCase.retreivedContentErr},
			MetadataRetriever: DummyMetadataRetriever{testCase.retrievedAnn, testCase.retreivedAnnErr},
		}

		m, err := combiner.GetCombinedModelForAnnotations(testCase.metadata, "some_platform_version")
		assert.Equal(t, testCase.expModel, m,
			fmt.Sprintf("Expected model: %v was not equal with the received one: %v \n", testCase.expModel, m))
		if testCase.expError == nil {
			assert.Equal(t, nil, err)
		} else {
			assert.Contains(t, err.Error(), testCase.expError.Error())
		}
	}
}

func TestGetAnnotations(t *testing.T) {

	tests := []struct {
		uuid           string
		address        utils.ApiURL
		dc             utils.Client
		expAnnotations []Annotation
		expError       error
	}{
		{
			"some_uuid",
			utils.ApiURL{"some_host", "some_endpoint"},
			dummyClient{
				statusCode: http.StatusNotFound,
			},
			[]Annotation(nil), //empty value for a slice
			nil,
		},
		{
			"some_uuid",
			utils.ApiURL{"some_host", "some_endpoint"},
			dummyClient{
				err: errors.New("some error"),
			},
			[]Annotation(nil), //empty value for a slice
			errors.New("some error"),
		},
		{
			"some_uuid",
			utils.ApiURL{"some_host", "some_endpoint"},
			dummyClient{
				statusCode: http.StatusOK,
				body:       "text that can't be unmarshalled",
			},
			[]Annotation(nil),
			errors.New("Could not unmarshall annotations for content with uuid=some_uuid"),
		},
		{
			"some_uuid",
			utils.ApiURL{"some_host", "some_endpoint"},
			dummyClient{
				statusCode: http.StatusOK,
				body:       `[{"predicate":"http://base-url/about","id":"http://base-url/80bec524-8c75-4d0f-92fa-abce3962d995","apiUrl":"http://base-url/80bec524-8c75-4d0f-92fa-abce3962d995","types":["http://base-url/core/Thing","http://base-url/concept/Concept","http://base-url/organisation/Organisation","http://base-url/company/Company","http://base-url/company/PublicCompany"],"leiCode":"leicode_id_1","prefLabel":"Barclays","factsetID":"factset-id1","tmeIDs":["tme_id1"],"uuids":["80bec524-8c75-4d0f-92fa-abce3962d995","factset-generated-uuid"],"platformVersion":"v1"},{"predicate":"http://base-url/isClassifiedBy","id":"http://base-url/271ee5f7-d808-497d-bed3-1b961953dedc","apiUrl":"http://base-url/271ee5f7-d808-497d-bed3-1b961953dedc","types":["http://base-url/core/Thing","http://base-url/concept/Concept","http://base-url/classification/Classification","http://base-url/Section"],"prefLabel":"Financials","tmeIDs":["tme_id_2"],"uuids":["271ee5f7-d808-497d-bed3-1b961953dedc"],"platformVersion":"v1"},{"predicate":"http://base-url/majorMentions","id":"http://base-url/a19d07d5-dc28-4c33-8745-a96f193df5cd","apiUrl":"http://base-url/a19d07d5-dc28-4c33-8745-a96f193df5cd","types":["http://base-url/core/Thing","http://base-url/concept/Concept","http://base-url/person/Person"],"prefLabel":"Jes Staley","tmeIDs":["tme_id_3"],"uuids":["a19d07d5-dc28-4c33-8745-a96f193df5cd"],"platformVersion":"v1"}]`,
			},
			[]Annotation{
				{
					Thing: Thing{
						ID:        "http://base-url/80bec524-8c75-4d0f-92fa-abce3962d995",
						PrefLabel: "Barclays",
						Types: []string{"http://base-url/core/Thing",
							"http://base-url/concept/Concept",
							"http://base-url/organisation/Organisation",
							"http://base-url/company/Company",
							"http://base-url/company/PublicCompany",
						},
						Predicate: "http://base-url/about",
						ApiUrl:    "http://base-url/80bec524-8c75-4d0f-92fa-abce3962d995",
						LeiCode:   "leicode_id_1",
						FactsetID: "factset-id1",
						TmeIDs:    []string{"tme_id1"},
						UUIDs: []string{"80bec524-8c75-4d0f-92fa-abce3962d995",
							"factset-generated-uuid"},
						PlatformVersion: "v1",
					},
				},
				{
					Thing: Thing{
						ID:        "http://base-url/271ee5f7-d808-497d-bed3-1b961953dedc",
						PrefLabel: "Financials",
						Types: []string{"http://base-url/core/Thing",
							"http://base-url/concept/Concept",
							"http://base-url/classification/Classification",
							"http://base-url/Section"},
						Predicate:       "http://base-url/isClassifiedBy",
						ApiUrl:          "http://base-url/271ee5f7-d808-497d-bed3-1b961953dedc",
						LeiCode:         "",
						FactsetID:       "",
						TmeIDs:          []string{"tme_id_2"},
						UUIDs:           []string{"271ee5f7-d808-497d-bed3-1b961953dedc"},
						PlatformVersion: "v1",
					},
				},
				{
					Thing: Thing{
						ID:        "http://base-url/a19d07d5-dc28-4c33-8745-a96f193df5cd",
						PrefLabel: "Jes Staley",
						Types: []string{"http://base-url/core/Thing",
							"http://base-url/concept/Concept",
							"http://base-url/person/Person"},
						Predicate:       "http://base-url/majorMentions",
						ApiUrl:          "http://base-url/a19d07d5-dc28-4c33-8745-a96f193df5cd",
						LeiCode:         "",
						FactsetID:       "",
						TmeIDs:          []string{"tme_id_3"},
						UUIDs:           []string{"a19d07d5-dc28-4c33-8745-a96f193df5cd"},
						PlatformVersion: "v1",
					},
				},
			},
			nil,
		},
	}

	for _, testCase := range tests {
		dr := dataRetriever{testCase.address, testCase.dc}
		ann, err := dr.getAnnotations(testCase.uuid, "some_platform_version")
		assert.Equal(t, testCase.expAnnotations, ann,
			fmt.Sprintf("Expected annotations: %v were not equal with received ones: %v \n", testCase.expAnnotations, ann))
		if testCase.expError == nil {
			assert.Equal(t, nil, err)
		} else {
			assert.Contains(t, err.Error(), testCase.expError.Error())
		}
	}
}

func TestGetContent(t *testing.T) {
	tests := []struct {
		uuid       string
		address    utils.ApiURL
		dc         utils.Client
		expContent ContentModel
		expError   error
	}{
		{
			"some_uuid",
			utils.ApiURL{"some_host", "some_endpoint"},
			dummyClient{
				statusCode: http.StatusNotFound,
			},
			ContentModel{},
			nil,
		},
		{
			"some_uuid",
			utils.ApiURL{"some_host", "some_endpoint"},
			dummyClient{
				err: errors.New("some error"),
			},
			ContentModel{},
			errors.New("some error"),
		},
		{
			"some_uuid",
			utils.ApiURL{"some_host", "some_endpoint"},
			dummyClient{
				statusCode: http.StatusOK,
				body:       "text that can't be unmarshalled",
			},
			ContentModel{},
			errors.New("Could not unmarshall content with uuid=some_uuid"),
		},
		{
			"622de808-3a7a-49bd-a7fb-2a33f64695be",
			utils.ApiURL{"some_host", "some_endpoint"},
			dummyClient{
				statusCode: http.StatusOK,
				body:       `{"uuid":"622de808-3a7a-49bd-a7fb-2a33f64695be","title":"Title","alternativeTitles":{"promotionalTitle":"Alternative title"},"type":null,"byline":"FT Reporters","brands":[{"id":"http://api.ft.com/things/40f636a3-5507-4311-9629-95376007cb7b"}],"identifiers":[{"authority":"FTCOM-METHODE_identifier","identifierValue":"53217c65-ecef-426e-a3ac-3787e2e62e87"}],"publishedDate":"2017-04-10T08:03:58.000Z","standfirst":"A simple line with an article summary","body":"<body>something relevant here<\/body>","description":null,"mediaType":null,"pixelWidth":null,"pixelHeight":null,"internalBinaryUrl":null,"externalBinaryUrl":null,"members":null,"mainImage":"2934de46-5240-4c7d-8576-f12ae12e4a37","standout":{"editorsChoice":false,"exclusive":false,"scoop":false},"comments":{"enabled":true},"copyright":null,"webUrl":null,"publishReference":"tid_unique_reference","lastModified":"2017-04-10T08:09:01.808Z","canBeSyndicated":"yes","firstPublishedDate":"2017-04-10T08:03:58.000Z","accessLevel":"subscribed","canBeDistributed":"yes"}`,
			},
			ContentModel{
				UUID:  "622de808-3a7a-49bd-a7fb-2a33f64695be",
				Title: "Title",
				Body:  "<body>something relevant here</body>",
				Identifiers: []Identifier{
					{
						Authority:       "FTCOM-METHODE_identifier",
						IdentifierValue: "53217c65-ecef-426e-a3ac-3787e2e62e87",
					},
				},
				PublishedDate:      "2017-04-10T08:03:58.000Z",
				LastModified:       "2017-04-10T08:09:01.808Z",
				FirstPublishedDate: "2017-04-10T08:03:58.000Z",
				MediaType:          "",
				MarkedDeleted:      false,
				Byline:             "FT Reporters",
				Standfirst:         "A simple line with an article summary",
				Description:        "",
				MainImage:          "2934de46-5240-4c7d-8576-f12ae12e4a37",
				PublishReference:   "tid_unique_reference",
			},
			nil,
		},
	}

	for _, testCase := range tests {
		dr := dataRetriever{testCase.address, testCase.dc}
		c, err := dr.getContent(testCase.uuid)

		assert.True(t, reflect.DeepEqual(testCase.expContent, c), fmt.Sprintf("Expected content: %v was not equal with received content: %v \n", testCase.expContent, c))
		if testCase.expError == nil {
			assert.Equal(t, nil, err)
		} else {
			assert.Contains(t, err.Error(), testCase.expError.Error())
		}
	}
}

type dummyClient struct {
	statusCode int
	body       string
	err        error
}

func (c dummyClient) Do(req *http.Request) (*http.Response, error) {

	resp := &http.Response{
		StatusCode: c.statusCode,
		Body:       ioutil.NopCloser(strings.NewReader(c.body)),
	}

	return resp, c.err
}

type DummyContentRetriever struct {
	c   ContentModel
	err error
}

func (r DummyContentRetriever) getContent(uuid string) (ContentModel, error) {
	return r.c, r.err
}

type DummyMetadataRetriever struct {
	ann []Annotation
	err error
}

func (r DummyMetadataRetriever) getAnnotations(uuid string, platformVersion string) ([]Annotation, error) {
	return r.ann, r.err
}
