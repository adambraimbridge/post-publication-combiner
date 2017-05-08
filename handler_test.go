package main

import (
	"errors"
	"fmt"
	"github.com/Financial-Times/post-publication-combiner/processor"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPostMessage(t *testing.T) {

	tests := []struct {
		contentType     string
		uuid            string
		platformVersion string
		err             error
		status          int
	}{
		{"article", "a78cf3ea-b221-46f8-8cbc-a61e5e454e88", "v1", nil, 200},
		{"video", "f1ce611d-caf8-4d82-b129-3f43b50a1fd0", "next-video", nil, 200},
		{"invalid", "0ead533c-751d-410b-aac2-4a203fd6e8ce", "", nil, 400},
		{"article", "invalid", "", nil, 400},
		{"article", "a78cf3ea-b221-46f8-8cbc-a61e5e454e88", "v1", errors.New("test error"), 500},
		{"article", "a78cf3ea-b221-46f8-8cbc-a61e5e454e88", "v1", processor.NotFoundError, 404},
		{"article", "a78cf3ea-b221-46f8-8cbc-a61e5e454e88", "v1", processor.InvalidContentTypeError, 422},
	}

	p := &DummyProcessor{t: t}

	rh := requestHandler{processor: p}
	servicesRouter := mux.NewRouter()
	servicesRouter.HandleFunc("/{content-type}/{id}", rh.postMessage).Methods("POST")

	r := http.NewServeMux()
	r.Handle("/", servicesRouter)

	server := httptest.NewServer(r)

	defer server.Close()

	for _, testCase := range tests {
		p.uuid = testCase.uuid
		p.platformVersion = testCase.platformVersion
		p.err = testCase.err
		resp, err := http.Post(fmt.Sprintf("%s/%s/%s", server.URL, testCase.contentType, testCase.uuid), "", nil)
		resp.Body.Close()
		assert.NoError(t, err)
		assert.Equal(t, testCase.status, resp.StatusCode)
	}
}

type DummyProcessor struct {
	t               *testing.T
	uuid            string
	platformVersion string
	err             error
}

func (p *DummyProcessor) ProcessMessages() {
	panic("implement me")
}

func (p *DummyProcessor) ForceMessagePublish(uuid string, platformVersion string) error {
	assert.Equal(p.t, p.uuid, uuid)
	assert.Equal(p.t, p.platformVersion, platformVersion)
	return p.err
}
