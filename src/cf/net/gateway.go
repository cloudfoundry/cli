package net

import (
	"bytes"
	"cf"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"runtime"
)

const INVALID_TOKEN_CODE = "GATEWAY INVALID TOKEN CODE"

type errorResponse struct {
	Code        string
	Description string
}

type errorHandler func(*http.Response) errorResponse

type tokenRefresher interface {
	RefreshAuthToken() (string, ApiResponse)
}

type Request struct {
	*http.Request
}

type Gateway struct {
	authenticator tokenRefresher
	errHandler    errorHandler
}

func newGateway(errHandler errorHandler) (gateway Gateway) {
	gateway.errHandler = errHandler
	return
}

func (gateway *Gateway) SetTokenRefresher(auth tokenRefresher) {
	gateway.authenticator = auth
}

func (gateway Gateway) GetResource(url, accessToken string, resource interface{}) (apiResponse ApiResponse) {
	request, apiResponse := gateway.NewRequest("GET", url, accessToken, nil)
	if apiResponse.IsNotSuccessful() {
		return
	}

	_, apiResponse = gateway.PerformRequestForJSONResponse(request, resource)
	return
}

func (gateway Gateway) CreateResource(url, accessToken string, body io.Reader) (apiResponse ApiResponse) {
	return gateway.createUpdateOrDeleteResource("POST", url, accessToken, body, nil)
}

func (gateway Gateway) CreateResourceForResponse(url, accessToken string, body io.Reader, resource interface{}) (apiResponse ApiResponse) {
	return gateway.createUpdateOrDeleteResource("POST", url, accessToken, body, resource)
}

func (gateway Gateway) UpdateResource(url, accessToken string, body io.Reader) (apiResponse ApiResponse) {
	return gateway.createUpdateOrDeleteResource("PUT", url, accessToken, body, nil)
}

func (gateway Gateway) UpdateResourceForResponse(url, accessToken string, body io.Reader, resource interface{}) (apiResponse ApiResponse) {
	return gateway.createUpdateOrDeleteResource("PUT", url, accessToken, body, resource)
}

func (gateway Gateway) DeleteResource(url, accessToken string) (apiResponse ApiResponse) {
	return gateway.createUpdateOrDeleteResource("DELETE", url, accessToken, nil, nil)
}

func (gateway Gateway) createUpdateOrDeleteResource(verb, url, accessToken string, body io.Reader, resource interface{}) (apiResponse ApiResponse) {
	request, apiResponse := gateway.NewRequest(verb, url, accessToken, body)
	if apiResponse.IsNotSuccessful() {
		return
	}

	if resource != nil {
		_, apiResponse = gateway.PerformRequestForJSONResponse(request, resource)
		return
	}

	return gateway.PerformRequest(request)
}

func (gateway Gateway) NewRequest(method, path, accessToken string, body io.Reader) (req *Request, apiResponse ApiResponse) {
	request, err := http.NewRequest(method, path, body)
	if err != nil {
		apiResponse = NewApiResponseWithError("Error building request", err)
		return
	}

	if accessToken != "" {
		request.Header.Set("Authorization", accessToken)
	}

	request.Header.Set("accept", "application/json")
	request.Header.Set("content-type", "application/json")
	request.Header.Set("User-Agent", "go-cli "+cf.Version+" / "+runtime.GOOS)
	req = &Request{request}
	return
}

func (gateway Gateway) PerformRequest(request *Request) (apiResponse ApiResponse) {
	_, apiResponse = gateway.doRequestHandlingAuth(request)
	return
}

func (gateway Gateway) PerformRequestForResponseBytes(request *Request) (bytes []byte, headers http.Header, apiResponse ApiResponse) {
	rawResponse, apiResponse := gateway.doRequestHandlingAuth(request)
	if apiResponse.IsNotSuccessful() {
		return
	}

	bytes, err := ioutil.ReadAll(rawResponse.Body)
	if err != nil {
		apiResponse = NewApiResponseWithError("Error reading response", err)
	}

	headers = rawResponse.Header
	return
}

func (gateway Gateway) PerformRequestForTextResponse(request *Request) (response string, headers http.Header, apiResponse ApiResponse) {
	bytes, headers, apiResponse := gateway.PerformRequestForResponseBytes(request)
	response = string(bytes)
	return
}

func (gateway Gateway) PerformRequestForJSONResponse(request *Request, response interface{}) (headers http.Header, apiResponse ApiResponse) {
	bytes, headers, apiResponse := gateway.PerformRequestForResponseBytes(request)
	if apiResponse.IsNotSuccessful() {
		return
	}

	err := json.Unmarshal(bytes, &response)
	if err != nil {
		apiResponse = NewApiResponseWithError("Invalid JSON response from server", err)
	}
	return
}

func (gateway Gateway) doRequestHandlingAuth(request *Request) (rawResponse *http.Response, apiResponse ApiResponse) {
	// copy body bytes for redoing request after an OAUTH refresh
	var bodyBytes []byte
	if request.Body != nil {
		bodyBytes, _ = ioutil.ReadAll(request.Body)
		request.Body = ioutil.NopCloser(bytes.NewReader(bodyBytes))
	}

	// perform request
	rawResponse, apiResponse = gateway.doRequestAndHandlerError(request)
	if apiResponse.IsSuccessful() || gateway.authenticator == nil {
		return
	}

	if apiResponse.ErrorCode != INVALID_TOKEN_CODE {
		return
	}

	// refresh the auth token
	newToken, apiResponse := gateway.authenticator.RefreshAuthToken()
	if apiResponse.IsNotSuccessful() {
		return
	}

	// reset the auth token and request body
	request.Header.Set("Authorization", newToken)
	if len(bodyBytes) > 0 {
		request.Body = ioutil.NopCloser(bytes.NewReader(bodyBytes))
	}

	// make the request again
	rawResponse, apiResponse = gateway.doRequestAndHandlerError(request)
	return
}

func (gateway Gateway) doRequestAndHandlerError(request *Request) (rawResponse *http.Response, apiResponse ApiResponse) {
	rawResponse, err := doRequest(request.Request)
	if err != nil {
		apiResponse = NewApiResponseWithError("Error performing request", err)
		return
	}

	if rawResponse.StatusCode > 299 {
		errorResponse := gateway.errHandler(rawResponse)
		message := fmt.Sprintf(
			"Server error, status code: %d, error code: %s, message: %s",
			rawResponse.StatusCode,
			errorResponse.Code,
			errorResponse.Description,
		)
		apiResponse = NewApiResponse(message, errorResponse.Code, rawResponse.StatusCode)
	}
	return
}
