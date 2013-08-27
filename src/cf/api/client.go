package api

import (
	term "cf/terminal"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"os"
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
	_, err = doRequest(request.Request)
	return
}

func PerformRequestAndParseResponse(request *AuthorizedRequest, response interface{}) (err error) {
	rawResponse, err := doRequest(request.Request)
	if err != nil {
		return
	}

	jsonBytes, err := ioutil.ReadAll(rawResponse.Body)
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

func doRequest(request *http.Request) (response *http.Response, err error) {
	client := newClient()

	if os.Getenv("TRACE") != "" {
		dumpedRequest, err := httputil.DumpRequest(request, true)
		if err != nil {
			fmt.Println("Error dumping request")
		} else {
			fmt.Printf("\n%s\n%s\n", term.Cyan("REQUEST:"), string(dumpedRequest))
		}
	}

	response, err = client.Do(request)

	if os.Getenv("TRACE") != "" {
		dumpedResponse, err := httputil.DumpResponse(response, true)
		if err != nil {
			fmt.Println("Error dumping response")
		} else {
			fmt.Printf("\n%s\n%s\n", term.Cyan("RESPONSE:"), string(dumpedResponse))
		}
	}

	if err != nil {
		err = errors.New(fmt.Sprintf("Error performing request: %s", err.Error()))
		return
	}

	if response.StatusCode > 299 {
		errorResponse := getErrorResponse(response)
		message := fmt.Sprintf("Server error, status code: %d, error code: %d, message: %s", response.StatusCode, errorResponse.Code, errorResponse.Description)
		err = errors.New(message)
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
