package net

import (
	"cf"
	"cf/errors"
	"crypto/tls"
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
	ASYNC_REQUEST_TIMEOUT    = 20 * time.Second
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
	RefreshAuthToken() (string, errors.Error)
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
	trustedCerts    []tls.Certificate
}

func newGateway(errHandler errorHandler) (gateway Gateway) {
	gateway.errHandler = errHandler
	gateway.PollingThrottle = DEFAULT_POLLING_THROTTLE
	return
}

func (gateway *Gateway) SetTokenRefresher(auth tokenRefresher) {
	gateway.authenticator = auth
}

func (gateway Gateway) GetResource(url, accessToken string, resource interface{}) (apiErr errors.Error) {
	request, apiErr := gateway.NewRequest("GET", url, accessToken, nil)
	if apiErr != nil {
		return
	}

	_, apiErr = gateway.PerformRequestForJSONResponse(request, resource)
	return
}

func (gateway Gateway) CreateResource(url, accessToken string, body io.ReadSeeker) (apiErr errors.Error) {
	return gateway.createUpdateOrDeleteResource("POST", url, accessToken, body, nil)
}

func (gateway Gateway) CreateResourceForResponse(url, accessToken string, body io.ReadSeeker, resource interface{}) (apiErr errors.Error) {
	return gateway.createUpdateOrDeleteResource("POST", url, accessToken, body, resource)
}

func (gateway Gateway) UpdateResource(url, accessToken string, body io.ReadSeeker) (apiErr errors.Error) {
	return gateway.createUpdateOrDeleteResource("PUT", url, accessToken, body, nil)
}

func (gateway Gateway) UpdateResourceForResponse(url, accessToken string, body io.ReadSeeker, resource interface{}) (apiErr errors.Error) {
	return gateway.createUpdateOrDeleteResource("PUT", url, accessToken, body, resource)
}

func (gateway Gateway) DeleteResource(url, accessToken string) (apiErr errors.Error) {
	return gateway.createUpdateOrDeleteResource("DELETE", url, accessToken, nil, &AsyncResponse{})
}

func (gateway Gateway) ListPaginatedResources(
	target string,
	accessToken string,
	path string,
	resource interface{},
	cb func(interface{}) bool) (apiErr errors.Error) {

	for path != "" {
		pagination := NewPaginatedResources(resource)
		apiErr = gateway.GetResource(fmt.Sprintf("%s%s", target, path), accessToken, &pagination)
		if apiErr != nil {
			return
		}

		resources, err := pagination.Resources()
		if err != nil {
			return errors.NewErrorWithError("Error parsing JSON", err)
		}

		for _, resource := range resources {
			if !cb(resource) {
				return
			}
		}

		path = pagination.NextURL
	}

	return
}

func (gateway Gateway) createUpdateOrDeleteResource(verb, url, accessToken string, body io.ReadSeeker, resource interface{}) (apiErr errors.Error) {
	request, apiErr := gateway.NewRequest(verb, url, accessToken, body)
	if apiErr != nil {
		return
	}

	if resource == nil {
		return gateway.PerformRequest(request)
	}

	if gateway.PollingEnabled {
		_, apiErr = gateway.PerformPollingRequestForJSONResponse(request, resource, ASYNC_REQUEST_TIMEOUT)
		return
	} else {
		_, apiErr = gateway.PerformRequestForJSONResponse(request, resource)
		return
	}
}

func (gateway Gateway) NewRequest(method, path, accessToken string, body io.ReadSeeker) (req *Request, apiErr errors.Error) {
	if body != nil {
		body.Seek(0, 0)
	}

	request, err := http.NewRequest(method, path, body)
	if err != nil {
		apiErr = errors.NewErrorWithError("Error building request", err)
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

func (gateway Gateway) PerformRequest(request *Request) (apiErr errors.Error) {
	_, apiErr = gateway.doRequestHandlingAuth(request)
	return
}

func (gateway Gateway) PerformRequestForResponse(request *Request) (rawResponse *http.Response, apiErr errors.Error) {
	return gateway.doRequestHandlingAuth(request)
}

func (gateway Gateway) PerformRequestForResponseBytes(request *Request) (bytes []byte, headers http.Header, rawResponse *http.Response, apiErr errors.Error) {
	rawResponse, apiErr = gateway.doRequestHandlingAuth(request)
	if apiErr != nil {
		return
	}

	bytes, err := ioutil.ReadAll(rawResponse.Body)
	if err != nil {
		apiErr = errors.NewErrorWithError("Error reading response", err)
	}

	headers = rawResponse.Header
	return
}

func (gateway Gateway) PerformRequestForTextResponse(request *Request) (response string, headers http.Header, apiErr errors.Error) {
	bytes, headers, _, apiErr := gateway.PerformRequestForResponseBytes(request)
	response = string(bytes)
	return
}

func (gateway Gateway) PerformRequestForJSONResponse(request *Request, response interface{}) (headers http.Header, apiErr errors.Error) {
	bytes, headers, rawResponse, apiErr := gateway.PerformRequestForResponseBytes(request)
	if apiErr != nil {
		return
	}

	if rawResponse.StatusCode > 203 || strings.TrimSpace(string(bytes)) == "" {
		return
	}

	err := json.Unmarshal(bytes, &response)
	if err != nil {
		apiErr = errors.NewErrorWithError("Invalid JSON response from server", err)
	}
	return
}

func (gateway Gateway) PerformPollingRequestForJSONResponse(request *Request, response interface{}, timeout time.Duration) (headers http.Header, apiErr errors.Error) {
	query := request.HttpReq.URL.Query()
	query.Add("async", "true")
	request.HttpReq.URL.RawQuery = query.Encode()

	bytes, headers, rawResponse, apiErr := gateway.PerformRequestForResponseBytes(request)
	if apiErr != nil {
		return
	}

	if rawResponse.StatusCode > 203 || strings.TrimSpace(string(bytes)) == "" {
		return
	}

	err := json.Unmarshal(bytes, &response)
	if err != nil {
		apiErr = errors.NewErrorWithError("Invalid JSON response from server", err)
		return
	}

	asyncResponse := &AsyncResponse{}

	err = json.Unmarshal(bytes, &asyncResponse)
	if err != nil {
		apiErr = errors.NewErrorWithError("Invalid async response from server", err)
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
	apiErr = gateway.waitForJob(jobUrl, request.HttpReq.Header.Get("Authorization"), timeout)

	return
}

func (gateway Gateway) waitForJob(jobUrl, accessToken string, timeout time.Duration) (apiErr errors.Error) {
	startTime := time.Now()
	for true {
		if time.Since(startTime) > timeout {
			apiErr = errors.NewErrorWithMessage("Error: timed out waiting for async job '%s' to finish", jobUrl)
			return
		}

		var request *Request
		request, apiErr = gateway.NewRequest("GET", jobUrl, accessToken, nil)
		response := &JobResponse{}

		_, apiErr = gateway.PerformRequestForJSONResponse(request, response)
		if apiErr != nil {
			return
		}

		switch response.Entity.Status {
		case JOB_FINISHED:
			return
		case JOB_FAILED:
			apiErr = errors.NewError("Job failed", JOB_FAILED)
			return
		}

		accessToken = request.HttpReq.Header.Get("Authorization")

		time.Sleep(gateway.PollingThrottle)
	}
	return
}

func (gateway Gateway) doRequestHandlingAuth(request *Request) (rawResponse *http.Response, apiErr errors.Error) {
	httpReq := request.HttpReq

	if request.SeekableBody != nil {
		httpReq.Body = ioutil.NopCloser(request.SeekableBody)
	}

	// perform request
	rawResponse, apiErr = gateway.doRequestAndHandlerError(request)
	if apiErr == nil || gateway.authenticator == nil {
		return
	}

	if apiErr != nil && apiErr.ErrorCode() != INVALID_TOKEN_CODE {
		return
	}

	// refresh the auth token
	newToken, apiErr := gateway.authenticator.RefreshAuthToken()
	if apiErr != nil {
		return
	}

	// reset the auth token and request body
	httpReq.Header.Set("Authorization", newToken)
	if request.SeekableBody != nil {
		request.SeekableBody.Seek(0, 0)
		httpReq.Body = ioutil.NopCloser(request.SeekableBody)
	}

	// make the request again
	rawResponse, apiErr = gateway.doRequestAndHandlerError(request)
	return
}

func (gateway Gateway) doRequestAndHandlerError(request *Request) (rawResponse *http.Response, apiErr errors.Error) {
	rawResponse, err := gateway.doRequest(request.HttpReq)
	if err != nil {
		apiErr = wrapSSLErrors(request.HttpReq.URL.Host, err)
		return
	}

	if rawResponse.StatusCode > 299 {
		errorResponse := gateway.errHandler(rawResponse)
		apiErr = errors.NewHttpError(rawResponse.StatusCode, errorResponse.ResponseHeader, errorResponse.ResponseBody, errorResponse.Code, errorResponse.Description)
	}

	return
}

func (gateway Gateway) doRequest(request *http.Request) (response *http.Response, err error) {
	httpClient := newHttpClient(gateway.trustedCerts)

	dumpRequest(request)

	response, err = httpClient.Do(request)
	if err != nil {
		return
	}

	dumpResponse(response)
	return
}

func (gateway *Gateway) SetTrustedCerts(certificates []tls.Certificate) {
	gateway.trustedCerts = certificates
}
