package combiner

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Financial-Times/post-publication-combiner/model"
	"github.com/Financial-Times/post-publication-combiner/utils"
	"net/http"
)

func GetCombinedModelForContent(address utils.ApiURL, client *http.Client, content model.ContentModel) (model.CombinedModel, error) {

	if content.UUID == "" {
		return model.CombinedModel{}, errors.New("Content has no UUID provided. Can't deduce annotations for it.")
	}

	ann, err := getAnnotations(content.UUID, address, client)
	if err != nil {
		return model.CombinedModel{}, err
	}

	return model.CombinedModel{
		UUID:       content.UUID,
		Content:    content,
		V1Metadata: ann,
	}, nil
}

func GetCombinedModelForAnnotations(docStoreAddress utils.ApiURL, publicAnnotationsAddress utils.ApiURL, client *http.Client, metadata model.Annotations) (model.CombinedModel, error) {

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
		d, err := getAnnotations(metadata.UUID, publicAnnotationsAddress, client)
		aCh <- annResponse{d, err}
	}()

	go func() {
		d, err := getContent(metadata.UUID, docStoreAddress, client)
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

func getAnnotations(uuid string, address utils.ApiURL, client *http.Client) ([]model.Annotation, error) {

	b, err := utils.ExecuteHTTPRequest(uuid, address, client)
	if err != nil {
		return nil, err
	}

	var things []model.Thing
	if err := json.Unmarshal(b, &things); err != nil {
		return nil, errors.New(fmt.Sprintf("Could not unmarshall annotations for content  with uuid=%v, error=%v", uuid, err.Error()))
	}

	var ann []model.Annotation
	for _, t := range things {
		ann = append(ann, model.Annotation{t})
	}

	return ann, nil
}

func getContent(uuid string, address utils.ApiURL, client *http.Client) (model.ContentModel, error) {

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
