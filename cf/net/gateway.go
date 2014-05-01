package net

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/cloudfoundry/cli/cf"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
	"time"
)

const (
	JOB_FINISHED             = "finished"
	JOB_FAILED               = "failed"
	DEFAULT_POLLING_THROTTLE = 5 * time.Second
	ASYNC_REQUEST_TIMEOUT    = 20 * time.Second
)

type JobResource struct {
	Entity struct {
		Status       string
		ErrorDetails struct {
			Description string
		} `json:"error_details"`
	}
}

type AsyncResource struct {
	Metadata struct {
		URL string
	}
}

type apiErrorHandler func(statusCode int, body []byte) error

type tokenRefresher interface {
	RefreshAuthToken() (string, error)
}

type Request struct {
	HttpReq      *http.Request
	SeekableBody io.ReadSeeker
}

type Gateway struct {
	authenticator   tokenRefresher
	errHandler      apiErrorHandler
	PollingEnabled  bool
	PollingThrottle time.Duration
	trustedCerts    []tls.Certificate
	config          configuration.Reader
	warnings        *[]string
}

func newGateway(errHandler apiErrorHandler, config configuration.Reader) (gateway Gateway) {
	gateway.errHandler = errHandler
	gateway.config = config
	gateway.PollingThrottle = DEFAULT_POLLING_THROTTLE
	gateway.warnings = &[]string{}
	return
}

func (gateway *Gateway) SetTokenRefresher(auth tokenRefresher) {
	gateway.authenticator = auth
}

func (gateway Gateway) GetResource(url string, resource interface{}) (err error) {
	request, err := gateway.NewRequest("GET", url, gateway.config.AccessToken(), nil)
	if err != nil {
		return
	}

	_, err = gateway.PerformRequestForJSONResponse(request, resource)
	return
}

func (gateway Gateway) CreateResourceFromStruct(url string, resource interface{}) error {
	bytes, err := json.Marshal(resource)
	if err != nil {
		return err
	}

	return gateway.CreateResource(url, strings.NewReader(string(bytes)))
}

func (gateway Gateway) UpdateResourceFromStruct(url string, resource interface{}) error {
	bytes, err := json.Marshal(resource)
	if err != nil {
		return err
	}

	return gateway.UpdateResource(url, strings.NewReader(string(bytes)))
}

func (gateway Gateway) CreateResource(url string, body io.ReadSeeker, resource ...interface{}) (apiErr error) {
	return gateway.createUpdateOrDeleteResource("POST", url, body, false, resource...)
}

func (gateway Gateway) UpdateResource(url string, body io.ReadSeeker, resource ...interface{}) (apiErr error) {
	return gateway.createUpdateOrDeleteResource("PUT", url, body, false, resource...)
}

func (gateway Gateway) UpdateResourceSync(url string, body io.ReadSeeker, resource ...interface{}) (apiErr error) {
	return gateway.createUpdateOrDeleteResource("PUT", url, body, true, resource...)
}

func (gateway Gateway) DeleteResource(url string) (apiErr error) {
	return gateway.createUpdateOrDeleteResource("DELETE", url, nil, false, &AsyncResource{})
}

func (gateway Gateway) ListPaginatedResources(target string,
	path string,
	resource interface{},
	cb func(interface{}) bool) (apiErr error) {

	for path != "" {
		pagination := NewPaginatedResources(resource)
		apiErr = gateway.GetResource(fmt.Sprintf("%s%s", target, path), &pagination)
		if apiErr != nil {
			return
		}

		resources, err := pagination.Resources()
		if err != nil {
			return errors.NewWithError("Error parsing JSON", err)
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

func (gateway Gateway) createUpdateOrDeleteResource(verb, url string, body io.ReadSeeker, sync bool, optionalResource ...interface{}) (apiErr error) {
	var resource interface{}
	if len(optionalResource) > 0 {
		resource = optionalResource[0]
	}

	request, apiErr := gateway.NewRequest(verb, url, gateway.config.AccessToken(), body)
	if apiErr != nil {
		return
	}

	if resource == nil {
		_, apiErr = gateway.PerformRequest(request)
		return
	}

	if gateway.PollingEnabled && !sync {
		_, apiErr = gateway.PerformPollingRequestForJSONResponse(request, resource, ASYNC_REQUEST_TIMEOUT)
		return
	} else {
		_, apiErr = gateway.PerformRequestForJSONResponse(request, resource)
		return
	}
}

func (gateway Gateway) NewRequest(method, path, accessToken string, body io.ReadSeeker) (req *Request, apiErr error) {
	if body != nil {
		body.Seek(0, 0)
	}

	request, err := http.NewRequest(method, path, body)
	if err != nil {
		apiErr = errors.NewWithError("Error building request", err)
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

func (gateway Gateway) PerformRequest(request *Request) (rawResponse *http.Response, apiErr error) {
	return gateway.doRequestHandlingAuth(request)
}

func (gateway Gateway) performRequestForResponseBytes(request *Request) (bytes []byte, headers http.Header, rawResponse *http.Response, apiErr error) {
	rawResponse, apiErr = gateway.doRequestHandlingAuth(request)
	if apiErr != nil {
		return
	}

	bytes, err := ioutil.ReadAll(rawResponse.Body)
	if err != nil {
		apiErr = errors.NewWithError("Error reading response", err)
	}

	headers = rawResponse.Header
	return
}

func (gateway Gateway) PerformRequestForTextResponse(request *Request) (response string, headers http.Header, apiErr error) {
	bytes, headers, _, apiErr := gateway.performRequestForResponseBytes(request)
	response = string(bytes)
	return
}

func (gateway Gateway) PerformRequestForJSONResponse(request *Request, response interface{}) (headers http.Header, apiErr error) {
	bytes, headers, rawResponse, apiErr := gateway.performRequestForResponseBytes(request)
	if apiErr != nil {
		return
	}

	if rawResponse.StatusCode > 203 || strings.TrimSpace(string(bytes)) == "" {
		return
	}

	err := json.Unmarshal(bytes, &response)
	if err != nil {
		apiErr = errors.NewWithError("Invalid JSON response from server", err)
	}
	return
}

func (gateway Gateway) PerformPollingRequestForJSONResponse(request *Request, response interface{}, timeout time.Duration) (headers http.Header, apiErr error) {
	query := request.HttpReq.URL.Query()
	query.Add("async", "true")
	request.HttpReq.URL.RawQuery = query.Encode()

	bytes, headers, rawResponse, apiErr := gateway.performRequestForResponseBytes(request)
	if apiErr != nil {
		return
	}

	if rawResponse.StatusCode > 203 || strings.TrimSpace(string(bytes)) == "" {
		return
	}

	err := json.Unmarshal(bytes, &response)
	if err != nil {
		apiErr = errors.NewWithError("Invalid JSON response from server", err)
		return
	}

	asyncResource := &AsyncResource{}
	err = json.Unmarshal(bytes, &asyncResource)
	if err != nil {
		apiErr = errors.NewWithError("Invalid async response from server", err)
		return
	}

	jobUrl := asyncResource.Metadata.URL
	if jobUrl == "" {
		return
	}

	if !strings.Contains(jobUrl, "/jobs/") {
		return
	}

	jobUrl = fmt.Sprintf("%s://%s%s", request.HttpReq.URL.Scheme, request.HttpReq.URL.Host, asyncResource.Metadata.URL)
	apiErr = gateway.waitForJob(jobUrl, request.HttpReq.Header.Get("Authorization"), timeout)

	return
}

func (gateway Gateway) Warnings() []string {
	return *gateway.warnings
}

func (gateway Gateway) waitForJob(jobUrl, accessToken string, timeout time.Duration) (err error) {
	startTime := time.Now()
	for true {
		if time.Since(startTime) > timeout {
			err = errors.NewWithFmt("Error: timed out waiting for async job '%s' to finish", jobUrl)
			return
		}

		var request *Request
		request, err = gateway.NewRequest("GET", jobUrl, accessToken, nil)
		response := &JobResource{}

		_, err = gateway.PerformRequestForJSONResponse(request, response)
		if err != nil {
			return
		}

		switch response.Entity.Status {
		case JOB_FINISHED:
			return
		case JOB_FAILED:
			err = errors.New(response.Entity.ErrorDetails.Description)
			return
		}

		accessToken = request.HttpReq.Header.Get("Authorization")

		time.Sleep(gateway.PollingThrottle)
	}
	return
}

func (gateway Gateway) doRequestHandlingAuth(request *Request) (rawResponse *http.Response, err error) {
	httpReq := request.HttpReq

	if request.SeekableBody != nil {
		httpReq.Body = ioutil.NopCloser(request.SeekableBody)
	}

	// perform request
	rawResponse, err = gateway.doRequestAndHandlerError(request)
	if err == nil || gateway.authenticator == nil {
		return
	}

	switch err.(type) {
	case *errors.InvalidTokenError:
		// refresh the auth token
		var newToken string
		newToken, err = gateway.authenticator.RefreshAuthToken()
		if err != nil {
			return
		}

		// reset the auth token and request body
		httpReq.Header.Set("Authorization", newToken)
		if request.SeekableBody != nil {
			request.SeekableBody.Seek(0, 0)
			httpReq.Body = ioutil.NopCloser(request.SeekableBody)
		}

		// make the request again
		rawResponse, err = gateway.doRequestAndHandlerError(request)
	}

	return
}

func (gateway Gateway) doRequestAndHandlerError(request *Request) (rawResponse *http.Response, err error) {
	rawResponse, err = gateway.doRequest(request.HttpReq)
	if err != nil {
		err = WrapNetworkErrors(request.HttpReq.URL.Host, err)
		return
	}

	if rawResponse.StatusCode > 299 {
		jsonBytes, _ := ioutil.ReadAll(rawResponse.Body)
		rawResponse.Body.Close()
		rawResponse.Body = ioutil.NopCloser(bytes.NewBuffer(jsonBytes))
		err = gateway.errHandler(rawResponse.StatusCode, jsonBytes)
	}

	return
}

func (gateway Gateway) doRequest(request *http.Request) (response *http.Response, err error) {
	httpClient := newHttpClient(gateway.trustedCerts, gateway.config.IsSSLDisabled())

	dumpRequest(request)

	response, err = httpClient.Do(request)
	if err != nil {
		return
	}

	dumpResponse(response)

	header := http.CanonicalHeaderKey("X-Cf-Warnings")
	raw_warnings := response.Header[header]
	for _, raw_warning := range raw_warnings {
		warning, _ := url.QueryUnescape(raw_warning)
		*gateway.warnings = append(*gateway.warnings, warning)
	}

	return
}

func (gateway *Gateway) SetTrustedCerts(certificates []tls.Certificate) {
	gateway.trustedCerts = certificates
}
