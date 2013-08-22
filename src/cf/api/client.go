package api

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

func newClient() *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	return &http.Client{Transport: tr}
}

func NewAuthorizedRequest(method, path, accessToken string, body io.Reader) (request *http.Request, err error) {
	request, err = http.NewRequest(method, path, body)
	if err != nil {
		return
	}
	request.Header.Set("Authorization", accessToken)
	return
}

func PerformRequest(request *http.Request) (err error) {
	client := newClient()
	request.Header.Set("accept", "application/json")

	rawResponse, err := client.Do(request)

	if err != nil {
		err = errors.New(fmt.Sprintf("Error performing request: %s", err.Error()))
		return
	}

	if rawResponse.StatusCode > 299 {
		err = errors.New(fmt.Sprintf("Server error, status code: %d", rawResponse.StatusCode))

	}
	return
}

func PerformRequestForBody(request *http.Request, response interface{}) (err error) {
	client := newClient()
	request.Header.Set("accept", "application/json")

	rawResponse, err := client.Do(request)

	if err != nil {
		err = errors.New(fmt.Sprintf("Error performing request: %s", err.Error()))
		return
	}

	if rawResponse.StatusCode > 299 {
		err = errors.New(fmt.Sprintf("Server error, status code: %d", rawResponse.StatusCode))
		return
	}

	jsonBytes, err := ioutil.ReadAll(rawResponse.Body)
	rawResponse.Body.Close()

	if err != nil {
		err = errors.New(fmt.Sprintf("Could not read response body: %s", err.Error()))
		return
	}

	err = json.Unmarshal(jsonBytes, &response)

	if err != nil {
		err = errors.New(fmt.Sprintf("Invalid JSON response from server: %s", err.Error()))
	}
	return
}
