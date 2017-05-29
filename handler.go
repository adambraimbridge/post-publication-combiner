package main

import (
	"github.com/Financial-Times/post-publication-combiner/processor"
	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/satori/go.uuid"
	"net/http"
)

const (
	idpathVar = "id"
)

type requestHandler struct {
	processor processor.Processor
}

func (handler *requestHandler) postMessage(writer http.ResponseWriter, request *http.Request) {
	uuid := mux.Vars(request)[idpathVar]

	defer request.Body.Close()

	if !isValidUUID(uuid) {
		logrus.Errorf("Invalid UUID %s", uuid)
		writer.WriteHeader(http.StatusBadRequest)
		return

	}

	err := handler.processor.ForceMessagePublish(uuid)
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

func isValidUUID(id string) bool {
	_, err := uuid.FromString(id)
	return err == nil
}
