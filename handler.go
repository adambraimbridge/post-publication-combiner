package main

import (
	"github.com/Financial-Times/post-publication-combiner/processor"
	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/satori/go.uuid"
	"net/http"
)

const (
	idpathVar          = "id"
	contentTypePathVar = "content-type"
	contentTypeArticle = "article"
	contentTypeVideo   = "video"
)

var contentTypes = []string{contentTypeArticle, contentTypeVideo}

type requestHandler struct {
	processor processor.Processor
}

func (handler *requestHandler) postMessage(writer http.ResponseWriter, request *http.Request) {
	uuid := mux.Vars(request)[idpathVar]
	contentType := mux.Vars(request)[contentTypePathVar]
	transactionID := request.Header.Get("X-Request-Id")

	defer request.Body.Close()

	if !isValidContentType(contentType) {
		logrus.Errorf("Invalid content type %s", contentType)
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	if !isValidUUID(uuid) {
		logrus.Errorf("Invalid UUID %s", uuid)
		writer.WriteHeader(http.StatusBadRequest)
		return

	}

	platform := processor.PlatformV1

	if contentType == contentTypeVideo {
		platform = processor.PlatformVideo
	}

	err := handler.processor.ForceMessagePublish(uuid, transactionID, platform)
	switch err {
	case nil:
		writer.WriteHeader(http.StatusOK)
	case processor.NotFoundError:
		writer.WriteHeader(http.StatusNotFound)
	case processor.InvalidContentTypeError:
		writer.WriteHeader(http.StatusUnprocessableEntity)
	default:
		writer.WriteHeader(http.StatusInternalServerError)
	}

}

func isValidContentType(contentType string) bool {
	for _, ct := range contentTypes {
		if contentType == ct {
			return true
		}
	}
	return false
}

func isValidUUID(id string) bool {
	_, err := uuid.FromString(id)
	return err == nil
}
