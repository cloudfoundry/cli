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
	RefreshAuthToken() (string, ApiStatus)
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

func (gateway Gateway) NewRequest(method, path, accessToken string, body io.Reader) (req *Request, apiStatus ApiStatus) {
	request, err := http.NewRequest(method, path, body)
	if err != nil {
		apiStatus = NewApiStatusWithError("Error building request", err)
		return
	}

	if accessToken != "" {
		request.Header.Set("Authorization", accessToken)
	}

	request.Header.Set("accept", "application/json")
	request.Header.Set("User-Agent", "go-cli "+cf.Version+" / "+runtime.GOOS)
	req = &Request{request}
	return
}

func (gateway Gateway) PerformRequest(request *Request) (apiStatus ApiStatus) {
	_, apiStatus = gateway.doRequestHandlingAuth(request)
	return
}

func (gateway Gateway) PerformRequestForResponseBytes(request *Request) (bytes []byte, headers http.Header, apiStatus ApiStatus) {
	rawResponse, apiStatus := gateway.doRequestHandlingAuth(request)
	if apiStatus.IsNotSuccessful() {
		return
	}

	bytes, err := ioutil.ReadAll(rawResponse.Body)
	if err != nil {
		apiStatus = NewApiStatusWithError("Error reading response", err)
	}
	return
}

func (gateway Gateway) PerformRequestForTextResponse(request *Request) (response string, headers http.Header, apiStatus ApiStatus) {
	bytes, headers, apiStatus := gateway.PerformRequestForResponseBytes(request)
	response = string(bytes)
	return
}

func (gateway Gateway) PerformRequestForJSONResponse(request *Request, response interface{}) (headers http.Header, apiStatus ApiStatus) {
	bytes, headers, apiStatus := gateway.PerformRequestForResponseBytes(request)
	if apiStatus.IsNotSuccessful() {
		return
	}

	err := json.Unmarshal(bytes, &response)
	if err != nil {
		apiStatus = NewApiStatusWithError("Invalid JSON response from server", err)
	}
	return
}

func (gateway Gateway) doRequestHandlingAuth(request *Request) (response *http.Response, apiStatus ApiStatus) {
	var bodyBytes []byte
	if request.Body != nil {
		bodyBytes, _ = ioutil.ReadAll(request.Body)
		request.Body = ioutil.NopCloser(bytes.NewReader(bodyBytes))
	}

	response, err := doRequest(request.Request)
	if err != nil {
		apiStatus = NewApiStatusWithError("Error performing request", err)
		return
	}

	if response.StatusCode > 299 {
		errorResponse := gateway.errHandler(response)
		message := fmt.Sprintf(
			"Server error, status code: %d, error code: %s, message: %s",
			response.StatusCode,
			errorResponse.Code,
			errorResponse.Description,
		)
		apiStatus = NewApiStatus(message, errorResponse.Code, response.StatusCode)
	}

	if apiStatus.IsSuccessful() || gateway.authenticator == nil {
		return
	}

	if apiStatus.ErrorCode != INVALID_TOKEN_CODE {
		return
	}

	// refresh the auth token
	newToken, apiStatus := gateway.authenticator.RefreshAuthToken()
	if apiStatus.IsNotSuccessful() {
		return
	}

	// reset the auth token and request body
	request.Header.Set("Authorization", newToken)
	if len(bodyBytes) > 0 {
		request.Body = ioutil.NopCloser(bytes.NewReader(bodyBytes))
	}

	// make the request again
	response, err = doRequest(request.Request)
	if err != nil {
		apiStatus = NewApiStatusWithError("Error performing request", err)
	}
	return
}
