package utils

import (
	"errors"
	"fmt"
	"github.com/Sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

type ApiURL struct {
	BaseURL  string
	Endpoint string
}

func ExecuteHTTPRequest(uuid string, apiUrl ApiURL, httpClient *http.Client) (b []byte, err error) {
	return executeHTTPRequest(uuid, apiUrl.BaseURL+apiUrl.Endpoint, httpClient)
}

func ExecuteSimpleHTTPRequest(urlStr string, httpClient *http.Client) (b []byte, err error) {
	return executeHTTPRequest("", urlStr, httpClient)
}

func executeHTTPRequest(uuid string, urlStr string, httpClient *http.Client) (b []byte, err error) {

	if uuid != "" {
		urlStr = strings.Replace(urlStr, "{uuid}", uuid, -1)
	}

	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error creating requests for url=%s, error=%v", urlStr, err))
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error executing requests for url=%s , error=%v", urlStr, err))
	}

	defer cleanUp(resp)

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("Connecting to %s was not successful. Status: %d", req.URL, resp.StatusCode))
	}

	b, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Could not parse payload from response for url=%s, error=%v", req.URL, err))
	}

	return b, nil
}

func cleanUp(resp *http.Response) {

	_, err := io.Copy(ioutil.Discard, resp.Body)
	if err != nil {
		logrus.Warningf("[%v]", err)
	}

	err = resp.Body.Close()
	if err != nil {
		logrus.Warningf("[%v]", err)
	}

}
