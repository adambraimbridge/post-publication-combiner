package utils

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

type dummyClient struct {
	statusCode int
	body       string
	err        error
}

func (c *dummyClient) Do(req *http.Request) (*http.Response, error) {

	resp := &http.Response{
		StatusCode: c.statusCode,
		Body:       ioutil.NopCloser(strings.NewReader(c.body)),
	}

	return resp, c.err
}

func TestExecuteHTTPRequest(t *testing.T) {

	tests := []struct {
		dc            dummyClient
		url           string
		expRespBody   []byte
		expRespStatus int
		expErrStr     string
	}{
		{
			dc: dummyClient{
				body: "hey",
			},
			url:           "one malformed:url",
			expRespBody:   nil,
			expRespStatus: -1,
			expErrStr:     "Error creating requests for url=one malformed:url, error=parse one malformed:url: first path segment in URL cannot contain colon",
		},
		{
			dc: dummyClient{
				statusCode: http.StatusInternalServerError,
				err:        errors.New("Some error"),
			},
			url:           "url",
			expRespBody:   nil,
			expRespStatus: http.StatusInternalServerError,
			expErrStr:     "Error executing requests for url=url, error=Some error",
		},
		{
			dc: dummyClient{
				statusCode: http.StatusNotFound,
				body:       "simple body",
				err:        nil,
			},
			url:           "url",
			expRespBody:   nil,
			expRespStatus: http.StatusNotFound,
			expErrStr:     fmt.Sprintf("Connecting to url was not successful. Status: %d", http.StatusNotFound),
		},
		{
			dc: dummyClient{
				statusCode: http.StatusOK,
				body:       "simple body",
				err:        nil,
			},
			url:           "url",
			expRespBody:   []byte("simple body"),
			expRespStatus: http.StatusOK,
			expErrStr:     "",
		},
	}

	for _, testCase := range tests {
		b, s, err := executeHTTPRequest(testCase.url, &testCase.dc)

		if err != nil {
			assert.Equal(t, testCase.expErrStr, err.Error(), fmt.Sprintf("Expected error %v not equal with received error %v", testCase.expErrStr, err))
		} else {
			assert.Equal(t, testCase.expErrStr, "", fmt.Sprintf("Expected error %v not equal with nil", testCase.expErrStr))
		}

		assert.Equal(t, testCase.expRespBody, b, fmt.Sprintf("Expected body %v not equal with received body %v", testCase.expRespBody, b))
		assert.Equal(t, testCase.expRespStatus, s, fmt.Sprintf("Expected status %v not equal with received status %v", testCase.expRespStatus, s))
	}
}
