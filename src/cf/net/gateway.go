package net

import (
	"cf"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"
)

const (
	INVALID_TOKEN_CODE       = "GATEWAY INVALID TOKEN CODE"
	JOB_FINISHED             = "finished"
	JOB_FAILED               = "failed"
	DEFAULT_POLLING_THROTTLE = 5 * time.Second
)

type JobEntity struct {
	Status string
}

type JobResponse struct {
	Entity JobEntity
}

type AsyncMetadata struct {
	Url string
}

type AsyncResponse struct {
	Metadata AsyncMetadata
}

type errorResponse struct {
	Code           string
	Description    string
	ResponseHeader string
	ResponseBody   string
}

type errorHandler func(*http.Response) errorResponse

type tokenRefresher interface {
	RefreshAuthToken() (string, ApiResponse)
}

type Request struct {
	HttpReq      *http.Request
	SeekableBody io.ReadSeeker
}

type Gateway struct {
	authenticator   tokenRefresher
	errHandler      errorHandler
	PollingEnabled  bool
	PollingThrottle time.Duration
}

func newGateway(errHandler errorHandler) (gateway Gateway) {
	gateway.errHandler = errHandler
	gateway.PollingThrottle = DEFAULT_POLLING_THROTTLE
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

func (gateway Gateway) CreateResource(url, accessToken string, body io.ReadSeeker) (apiResponse ApiResponse) {
	return gateway.createUpdateOrDeleteResource("POST", url, accessToken, body, nil)
}

func (gateway Gateway) CreateResourceForResponse(url, accessToken string, body io.ReadSeeker, resource interface{}) (apiResponse ApiResponse) {
	return gateway.createUpdateOrDeleteResource("POST", url, accessToken, body, resource)
}

func (gateway Gateway) UpdateResource(url, accessToken string, body io.ReadSeeker) (apiResponse ApiResponse) {
	return gateway.createUpdateOrDeleteResource("PUT", url, accessToken, body, nil)
}

func (gateway Gateway) UpdateResourceForResponse(url, accessToken string, body io.ReadSeeker, resource interface{}) (apiResponse ApiResponse) {
	return gateway.createUpdateOrDeleteResource("PUT", url, accessToken, body, resource)
}

func (gateway Gateway) DeleteResource(url, accessToken string) (apiResponse ApiResponse) {
	return gateway.createUpdateOrDeleteResource("DELETE", url, accessToken, nil, &AsyncResponse{})
}

func (gateway Gateway) createUpdateOrDeleteResource(verb, url, accessToken string, body io.ReadSeeker, resource interface{}) (apiResponse ApiResponse) {
	request, apiResponse := gateway.NewRequest(verb, url, accessToken, body)
	if apiResponse.IsNotSuccessful() {
		return
	}

	if resource == nil {
		return gateway.PerformRequest(request)
	}

	if gateway.PollingEnabled {
		_, apiResponse = gateway.PerformPollingRequestForJSONResponse(request, resource)
		return
	} else {
		_, apiResponse = gateway.PerformRequestForJSONResponse(request, resource)
		return
	}

}

func (gateway Gateway) NewRequest(method, path, accessToken string, body io.ReadSeeker) (req *Request, apiResponse ApiResponse) {
	if body != nil {
		body.Seek(0, 0)
	}

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

	if body != nil {
		switch v := body.(type) {
		case *os.File:
			fileStats, err := v.Stat()
			if err != nil {
				break
			}
			request.ContentLength = fileStats.Size()
		}
	}

	req = &Request{HttpReq: request, SeekableBody: body}
	return
}

func (gateway Gateway) PerformRequest(request *Request) (apiResponse ApiResponse) {
	_, apiResponse = gateway.doRequestHandlingAuth(request)
	return
}

func (gateway Gateway) PerformRequestForResponse(request *Request) (rawResponse *http.Response, apiResponse ApiResponse) {
	return gateway.doRequestHandlingAuth(request)
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

	if apiResponse.StatusCode > 203 || strings.TrimSpace(string(bytes)) == "" {
		return
	}

	err := json.Unmarshal(bytes, &response)
	if err != nil {
		apiResponse = NewApiResponseWithError("Invalid JSON response from server", err)
	}
	return
}

func (gateway Gateway) PerformPollingRequestForJSONResponse(request *Request, response interface{}) (headers http.Header, apiResponse ApiResponse) {
	query := request.HttpReq.URL.Query()
	query.Add("async", "true")
	request.HttpReq.URL.RawQuery = query.Encode()

	bytes, headers, apiResponse := gateway.PerformRequestForResponseBytes(request)
	if apiResponse.IsNotSuccessful() {
		return
	}

	if apiResponse.StatusCode > 203 || strings.TrimSpace(string(bytes)) == "" {
		return
	}

	err := json.Unmarshal(bytes, &response)
	if err != nil {
		apiResponse = NewApiResponseWithError("Invalid JSON response from server", err)
		return
	}

	asyncResponse := &AsyncResponse{}

	err = json.Unmarshal(bytes, &asyncResponse)
	if err != nil {
		apiResponse = NewApiResponseWithError("Invalid async response from server", err)
		return
	}

	jobUrl := asyncResponse.Metadata.Url
	if jobUrl == "" {
		return
	}

	if !strings.Contains(jobUrl, "/jobs/") {
		return
	}

	jobUrl = fmt.Sprintf("%s://%s%s", request.HttpReq.URL.Scheme, request.HttpReq.URL.Host, asyncResponse.Metadata.Url)
	apiResponse = gateway.waitForJob(jobUrl, request.HttpReq.Header.Get("Authorization"))

	return
}

func (gateway Gateway) waitForJob(jobUrl, accessToken string) (apiResponse ApiResponse) {
	for true {
		var request *Request
		request, apiResponse = gateway.NewRequest("GET", jobUrl, accessToken, nil)
		response := &JobResponse{}

		_, apiResponse = gateway.PerformRequestForJSONResponse(request, response)
		if apiResponse.IsNotSuccessful() {
			return
		}

		switch response.Entity.Status {
		case JOB_FINISHED:
			return
		case JOB_FAILED:
			apiResponse = NewApiResponse("Internal Server Error", "", 500)
			return
		}

		accessToken = request.HttpReq.Header.Get("Authorization")

		time.Sleep(gateway.PollingThrottle)
	}
	return
}

func (gateway Gateway) doRequestHandlingAuth(request *Request) (rawResponse *http.Response, apiResponse ApiResponse) {
	httpReq := request.HttpReq

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
	httpReq.Header.Set("Authorization", newToken)
	if request.SeekableBody != nil {
		request.SeekableBody.Seek(0, 0)
		httpReq.Body = ioutil.NopCloser(request.SeekableBody)
	}

	// make the request again
	rawResponse, apiResponse = gateway.doRequestAndHandlerError(request)
	return
}

func (gateway Gateway) doRequestAndHandlerError(request *Request) (rawResponse *http.Response, apiResponse ApiResponse) {
	rawResponse, err := doRequest(request.HttpReq)
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
		apiResponse = NewApiResponseWithHttpError(message, errorResponse.Code, rawResponse.StatusCode, errorResponse.ResponseHeader, errorResponse.ResponseBody)
	} else {
		apiResponse = NewApiResponseWithStatusCode(rawResponse.StatusCode)
	}
	return
}
