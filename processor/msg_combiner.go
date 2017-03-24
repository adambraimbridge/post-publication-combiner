package processor

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Financial-Times/post-publication-combiner/model"
	"github.com/Financial-Times/post-publication-combiner/utils"
	"net/http"
)

type MessageCombiner struct {
	docStoreApi          utils.ApiURL
	publicAnnotationsApi utils.ApiURL
	httpClient           http.Client
}

func NewMessageCombiner(docStoreApiBaseURL string, docStoreApiEndpoint string, publicAnnotationsApiBaseURL string, publicAnnotationsApiEndpoint string, httpClient http.Client) *MessageCombiner {
	return &MessageCombiner{
		docStoreApi:          utils.ApiURL{docStoreApiBaseURL, docStoreApiEndpoint},
		publicAnnotationsApi: utils.ApiURL{publicAnnotationsApiBaseURL, publicAnnotationsApiEndpoint},
		httpClient:           httpClient,
	}
}

func (mc *MessageCombiner) enrichWithContent(metadata model.Annotations) (model.CombinedModel, error) {

	// even though we have the annotations from the kafka queue, we still need to read them from the DB to obtain the TME ids
	// Note: more detailed checks could be introduced at this point, if we want to make sure we have the same annotations returned

	if metadata.UUID == "" {
		return model.CombinedModel{}, errors.New("Annotations have no UUID referenced. Can't deduce content for it.")
	}

	type annResponse struct {
		ann []model.Annotation
		err error
	}
	type cResponse struct {
		c   model.ContentModel
		err error
	}

	cCh := make(chan cResponse)
	aCh := make(chan annResponse)

	go func() {
		d, err := getAnnotations(metadata.UUID, mc.publicAnnotationsApi, mc.httpClient)
		aCh <- annResponse{d, err}
	}()

	go func() {
		d, err := getContent(metadata.UUID, mc.docStoreApi, mc.httpClient)
		cCh <- cResponse{d, err}
	}()

	a := <-aCh
	if a.err != nil {
		return model.CombinedModel{}, a.err
	}

	c := <-cCh
	if c.err != nil {
		return model.CombinedModel{}, c.err
	}

	return model.CombinedModel{
		UUID:       metadata.UUID,
		Content:    c.c,
		V1Metadata: a.ann,
	}, nil
}

func (mc *MessageCombiner) enrichWithAnnotations(content model.ContentModel) (model.CombinedModel, error) {

	if content.UUID == "" {
		return model.CombinedModel{}, errors.New("Content has no UUID provided. Can't deduce annotations for it.")
	}

	ann, err := getAnnotations(content.UUID, mc.publicAnnotationsApi, mc.httpClient)
	if err != nil {
		return model.CombinedModel{}, err
	}

	return model.CombinedModel{
		UUID:       content.UUID,
		Content:    content,
		V1Metadata: ann,
	}, nil
}

func getAnnotations(uuid string, address utils.ApiURL, client http.Client) ([]model.Annotation, error) {

	b, err := utils.ExecuteHTTPRequest(uuid, address, client)
	if err != nil {
		return nil, err
	}

	var ann []model.Annotation
	if err := json.Unmarshal(b, &ann); err != nil {
		return nil, errors.New(fmt.Sprintf("Could not unmarshall annotations for content  with uuid=%v, error=%v", uuid, err.Error()))
	}

	return ann, nil
}

func getContent(uuid string, address utils.ApiURL, client http.Client) (model.ContentModel, error) {

	b, err := utils.ExecuteHTTPRequest(uuid, address, client)
	if err != nil {
		return model.ContentModel{}, err
	}

	var c model.ContentModel
	if err := json.Unmarshal(b, &c); err != nil {
		return model.ContentModel{}, errors.New(fmt.Sprintf("Could not unmarshall content with uuid=%v, error=%v", uuid, err.Error()))
	}

	return c, nil
}
