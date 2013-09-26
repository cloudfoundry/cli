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
	RefreshAuthToken() (string, *ApiError)
}

type Request struct {
	*http.Request
}

type Gateway struct {
	authenticator tokenRefresher
	errHandler    errorHandler
}

func newGateway(auth tokenRefresher, errHandler errorHandler) (gateway Gateway) {
	gateway.authenticator = auth
	gateway.errHandler = errHandler
	return
}

func (gateway Gateway) NewRequest(method, path, accessToken string, body io.Reader) (req *Request, apiErr *ApiError) {
	request, err := http.NewRequest(method, path, body)
	if err != nil {
		apiErr = NewApiErrorWithError("Error building request", err)
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

func (gateway Gateway) PerformRequest(request *Request) (apiErr *ApiError) {
	_, apiErr = gateway.doRequestHandlingAuth(request)
	return
}

func (gateway Gateway) PerformRequestForResponseBytes(request *Request) (bytes []byte, headers http.Header, apiErr *ApiError) {
	rawResponse, apiErr := gateway.doRequestHandlingAuth(request)
	if apiErr != nil {
		return
	}

	bytes, err := ioutil.ReadAll(rawResponse.Body)
	if err != nil {
		apiErr = NewApiErrorWithError("Error reading response", err)
	}
	return
}

func (gateway Gateway) PerformRequestForTextResponse(request *Request) (response string, headers http.Header, apiErr *ApiError) {
	bytes, headers, apiErr := gateway.PerformRequestForResponseBytes(request)
	response = string(bytes)
	return
}

func (gateway Gateway) PerformRequestForJSONResponse(request *Request, response interface{}) (headers http.Header, apiErr *ApiError) {
	bytes, headers, apiErr := gateway.PerformRequestForResponseBytes(request)
	if apiErr != nil {
		return
	}

	err := json.Unmarshal(bytes, &response)
	if err != nil {
		apiErr = NewApiErrorWithError("Invalid JSON response from server", err)
	}
	return
}

func (gateway Gateway) doRequestHandlingAuth(request *Request) (response *http.Response, apiErr *ApiError) {
	var bodyBytes []byte
	if request.Body != nil {
		bodyBytes, _ = ioutil.ReadAll(request.Body)
		request.Body = ioutil.NopCloser(bytes.NewReader(bodyBytes))
	}

	response, err := doRequest(request.Request)

	if err != nil && response == nil {
		apiErr = NewApiErrorWithError("Error performing request", err)
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
		apiErr = NewApiError(message, errorResponse.Code, response.StatusCode)
	}

	if apiErr == nil || gateway.authenticator == nil {
		return
	}

	if apiErr.ErrorCode == INVALID_TOKEN_CODE {
		newToken, apiErr := gateway.authenticator.RefreshAuthToken()
		if apiErr == nil {
			request.Header.Set("Authorization", newToken)
			if len(bodyBytes) > 0 {
				request.Body = ioutil.NopCloser(bytes.NewReader(bodyBytes))
			}

			response, err = doRequest(request.Request)
			if err != nil {
				apiErr = NewApiErrorWithError("Error performing request", err)
			}
			return response, apiErr
		}
	}

	return
}
