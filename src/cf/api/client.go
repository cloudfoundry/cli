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

type AuthorizedRequest struct {
	*http.Request
}

type errorResponse struct {
	Code        int
	Description string
}

func newClient() *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	return &http.Client{Transport: tr}
}

func NewAuthorizedRequest(method, path, accessToken string, body io.Reader) (*AuthorizedRequest, error) {
	request, err := http.NewRequest(method, path, body)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", accessToken)
	request.Header.Set("accept", "application/json")

	return &AuthorizedRequest{request}, err
}

func PerformRequest(request *AuthorizedRequest) (err error) {
	client := newClient()

	rawResponse, err := client.Do(request.Request)

	if err != nil {
		err = errors.New(fmt.Sprintf("Error performing request: %s", err.Error()))
		return
	}

	if rawResponse.StatusCode > 299 {
		errorResponse := getErrorResponse(rawResponse)
		message := fmt.Sprintf("Server error, status code: %d, message: %s", rawResponse.StatusCode, errorResponse.Description)
		err = errors.New(message)
	}
	return
}

func PerformRequestForBody(request *AuthorizedRequest, response interface{}) (err error) {
	client := newClient()

	rawResponse, err := client.Do(request.Request)

	if err != nil {
		err = errors.New(fmt.Sprintf("Error performing request: %s", err.Error()))
		return
	}

	if rawResponse.StatusCode > 299 {
		errorResponse := getErrorResponse(rawResponse)
		message := fmt.Sprintf("Server error, status code: %d, message: %s", rawResponse.StatusCode, errorResponse.Description)
		err = errors.New(message)
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

func getErrorResponse(response *http.Response) (eR errorResponse) {
	jsonBytes, _ := ioutil.ReadAll(response.Body)
	response.Body.Close()

	eR = errorResponse{}
	_ = json.Unmarshal(jsonBytes, &eR)
	return
}
