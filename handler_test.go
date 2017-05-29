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
		uuid   string
		err    error
		status int
	}{
		{"a78cf3ea-b221-46f8-8cbc-a61e5e454e88", nil, 200},
		{"invalid", nil, 400},
		{"a78cf3ea-b221-46f8-8cbc-a61e5e454e88", errors.New("test error"), 500},
		{"a78cf3ea-b221-46f8-8cbc-a61e5e454e88", processor.NotFoundError, 404},
		{"a78cf3ea-b221-46f8-8cbc-a61e5e454e88", processor.InvalidContentTypeError, 422},
	}

	p := &DummyProcessor{t: t}

	rh := requestHandler{processor: p}
	servicesRouter := mux.NewRouter()
	servicesRouter.HandleFunc("/{id}", rh.postMessage).Methods("POST")

	r := http.NewServeMux()
	r.Handle("/", servicesRouter)

	server := httptest.NewServer(r)

	defer server.Close()

	for _, testCase := range tests {
		p.uuid = testCase.uuid
		p.err = testCase.err
		resp, err := http.Post(fmt.Sprintf("%s/%s", server.URL, testCase.uuid), "", nil)
		resp.Body.Close()
		assert.NoError(t, err)
		assert.Equal(t, testCase.status, resp.StatusCode)
	}
}

type DummyProcessor struct {
	t    *testing.T
	uuid string
	err  error
}

func (p *DummyProcessor) ProcessMessages() {
	panic("implement me")
}

func (p *DummyProcessor) ForceMessagePublish(uuid string) error {
	assert.Equal(p.t, p.uuid, uuid)
	return p.err
}
