package processor

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Financial-Times/post-publication-combiner/model"
	"github.com/Financial-Times/post-publication-combiner/utils"
	"net/http"
	"strings"
)

type DataCombinerI interface {
	GetCombinedModelForContent(content model.ContentModel, platformVersion string) (model.CombinedModel, error)
	GetCombinedModelForAnnotations(metadata model.Annotations, platformVersion string) (model.CombinedModel, error)
	GetCombinedModel(uuid string, platformVersion string) (model.CombinedModel, error)
}

type DataCombiner struct {
	ContentRetriever  contentRetrieverI
	MetadataRetriever metadataRetrieverI
}

type contentRetrieverI interface {
	getContent(uuid string) (model.ContentModel, error)
}

type metadataRetrieverI interface {
	getAnnotations(uuid string, platfromVersion string) ([]model.Annotation, error)
}

type dataRetriever struct {
	Address utils.ApiURL
	client  utils.Client
}

func (dc DataCombiner) GetCombinedModelForContent(content model.ContentModel, platformVersion string) (model.CombinedModel, error) {

	if content.UUID == "" {
		return model.CombinedModel{}, errors.New("Content has no UUID provided. Can't deduce annotations for it.")
	}

	ann, err := dc.MetadataRetriever.getAnnotations(content.UUID, platformVersion)
	if err != nil {
		return model.CombinedModel{}, err
	}

	return model.CombinedModel{
		UUID:     content.UUID,
		Content:  content,
		Metadata: ann,
	}, nil
}

func (dc DataCombiner) GetCombinedModelForAnnotations(metadata model.Annotations, platformVersion string) (model.CombinedModel, error) {

	if metadata.UUID == "" {
		return model.CombinedModel{}, errors.New("Annotations have no UUID referenced. Can't deduce content for it.")
	}

	return dc.GetCombinedModel(metadata.UUID, platformVersion)
}

func (dc DataCombiner) GetCombinedModel(uuid string, platformVersion string) (model.CombinedModel, error) {
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
		d, err := dc.MetadataRetriever.getAnnotations(uuid, platformVersion)
		aCh <- annResponse{d, err}
	}()

	go func() {
		d, err := dc.ContentRetriever.getContent(uuid)
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
		UUID:     uuid,
		Content:  c.c,
		Metadata: a.ann,
	}, nil
}

func (dr dataRetriever) getAnnotations(uuid string, platformVersion string) ([]model.Annotation, error) {

	var ann []model.Annotation

	if platformVersion != "" {
		dr.Address.Endpoint = strings.Replace(dr.Address.Endpoint, "{platformVersion}", platformVersion, -1)
	}

	b, status, err := utils.ExecuteHTTPRequest(uuid, dr.Address, dr.client)

	if status == http.StatusNotFound {
		return ann, nil
	}

	if err != nil {
		return ann, err
	}

	var things []model.Thing
	if err := json.Unmarshal(b, &things); err != nil {
		return ann, fmt.Errorf("Could not unmarshall annotations for content with uuid=%v, error=%v", uuid, err.Error())
	}
	for _, t := range things {
		ann = append(ann, model.Annotation{t})
	}

	return ann, nil
}

func (dr dataRetriever) getContent(uuid string) (model.ContentModel, error) {

	var c model.ContentModel
	b, status, err := utils.ExecuteHTTPRequest(uuid, dr.Address, dr.client)

	if status == http.StatusNotFound {
		return c, nil
	}

	if err != nil {
		return c, err
	}

	if err := json.Unmarshal(b, &c); err != nil {
		return c, fmt.Errorf("Could not unmarshall content with uuid=%v, error=%v", uuid, err.Error())
	}

	return c, nil
}
