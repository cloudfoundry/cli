package net

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/cloudfoundry/cli/cf"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/terminal"
)

const (
	JOB_FINISHED             = "finished"
	JOB_FAILED               = "failed"
	DEFAULT_POLLING_THROTTLE = 5 * time.Second
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
	config          core_config.Reader
	warnings        *[]string
	Clock           func() time.Time
	transport       *http.Transport
	ui              terminal.UI
}

func newGateway(errHandler apiErrorHandler, config core_config.Reader, ui terminal.UI) (gateway Gateway) {
	gateway.errHandler = errHandler
	gateway.config = config
	gateway.PollingThrottle = DEFAULT_POLLING_THROTTLE
	gateway.warnings = &[]string{}
	gateway.Clock = time.Now
	gateway.ui = ui

	return
}

func (gateway *Gateway) AsyncTimeout() time.Duration {
	if gateway.config.AsyncTimeout() > 0 {
		return time.Duration(gateway.config.AsyncTimeout()) * time.Minute
	}

	return 0
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

func (gateway Gateway) CreateResourceFromStruct(endpoint, url string, resource interface{}) error {
	bytes, err := json.Marshal(resource)
	if err != nil {
		return err
	}

	return gateway.CreateResource(endpoint, url, strings.NewReader(string(bytes)))
}

func (gateway Gateway) UpdateResourceFromStruct(endpoint, apiUrl string, resource interface{}) error {
	bytes, err := json.Marshal(resource)
	if err != nil {
		return err
	}

	return gateway.UpdateResource(endpoint, apiUrl, strings.NewReader(string(bytes)))
}

func (gateway Gateway) CreateResource(endpoint, apiUrl string, body io.ReadSeeker, resource ...interface{}) (apiErr error) {
	return gateway.createUpdateOrDeleteResource("POST", endpoint, apiUrl, body, false, resource...)
}

func (gateway Gateway) UpdateResource(endpoint, apiUrl string, body io.ReadSeeker, resource ...interface{}) (apiErr error) {
	return gateway.createUpdateOrDeleteResource("PUT", endpoint, apiUrl, body, false, resource...)
}

func (gateway Gateway) UpdateResourceSync(endpoint, apiUrl string, body io.ReadSeeker, resource ...interface{}) (apiErr error) {
	return gateway.createUpdateOrDeleteResource("PUT", endpoint, apiUrl, body, true, resource...)
}

func (gateway Gateway) DeleteResourceSynchronously(endpoint, apiUrl string) (apiErr error) {
	return gateway.createUpdateOrDeleteResource("DELETE", endpoint, apiUrl, nil, true, &AsyncResource{})
}

func (gateway Gateway) DeleteResource(endpoint, apiUrl string) (apiErr error) {
	return gateway.createUpdateOrDeleteResource("DELETE", endpoint, apiUrl, nil, false, &AsyncResource{})
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
			return errors.NewWithError(T("Error parsing JSON"), err)
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

func (gateway Gateway) createUpdateOrDeleteResource(verb, endpoint, apiUrl string, body io.ReadSeeker, sync bool, optionalResource ...interface{}) (apiErr error) {
	var resource interface{}
	if len(optionalResource) > 0 {
		resource = optionalResource[0]
	}

	request, apiErr := gateway.NewRequest(verb, endpoint+apiUrl, gateway.config.AccessToken(), body)
	if apiErr != nil {
		return
	}

	if resource == nil {
		_, apiErr = gateway.PerformRequest(request)
		return
	}

	if gateway.PollingEnabled && !sync {
		_, apiErr = gateway.PerformPollingRequestForJSONResponse(endpoint, request, resource, gateway.AsyncTimeout())
		return
	} else {
		_, apiErr = gateway.PerformRequestForJSONResponse(request, resource)
		return
	}

}

func (gateway Gateway) newRequest(request *http.Request, accessToken string, body io.ReadSeeker) (*Request, error) {
	if accessToken != "" {
		request.Header.Set("Authorization", accessToken)
	}

	request.Header.Set("accept", "application/json")
	request.Header.Set("content-type", "application/json")
	request.Header.Set("User-Agent", "go-cli "+cf.Version+" / "+runtime.GOOS)
	return &Request{HttpReq: request, SeekableBody: body}, nil
}

func (gateway Gateway) NewRequestForFile(method, fullUrl, accessToken string, body *os.File) (req *Request, apiErr error) {
	progressReader := NewProgressReader(body, gateway.ui, 5*time.Second)
	progressReader.Seek(0, 0)
	fileStats, err := body.Stat()

	if err != nil {
		apiErr = errors.NewWithError(T("Error getting file info"), err)
		return
	}

	request, err := http.NewRequest(method, fullUrl, progressReader)
	if err != nil {
		apiErr = errors.NewWithError(T("Error building request"), err)
		return
	}

	fileSize := fileStats.Size()
	progressReader.SetTotalSize(fileSize)
	request.ContentLength = fileSize

	if err != nil {
		apiErr = errors.NewWithError(T("Error building request"), err)
		return
	}

	return gateway.newRequest(request, accessToken, progressReader)
}

func (gateway Gateway) NewRequest(method, path, accessToken string, body io.ReadSeeker) (req *Request, apiErr error) {
	request, err := http.NewRequest(method, path, body)
	if err != nil {
		apiErr = errors.NewWithError(T("Error building request"), err)
		return
	}
	return gateway.newRequest(request, accessToken, body)
}

func (gateway Gateway) PerformRequest(request *Request) (rawResponse *http.Response, apiErr error) {
	return gateway.doRequestHandlingAuth(request)
}

func (gateway Gateway) performRequestForResponseBytes(request *Request) (bytes []byte, headers http.Header, rawResponse *http.Response, apiErr error) {
	rawResponse, apiErr = gateway.doRequestHandlingAuth(request)
	if apiErr != nil {
		return
	}
	defer rawResponse.Body.Close()

	bytes, err := ioutil.ReadAll(rawResponse.Body)
	if err != nil {
		apiErr = errors.NewWithError(T("Error reading response"), err)
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
		apiErr = errors.NewWithError(T("Invalid JSON response from server"), err)
	}
	return
}

func (gateway Gateway) PerformPollingRequestForJSONResponse(endpoint string, request *Request, response interface{}, timeout time.Duration) (headers http.Header, apiErr error) {
	query := request.HttpReq.URL.Query()
	query.Add("async", "true")
	request.HttpReq.URL.RawQuery = query.Encode()

	bytes, headers, rawResponse, apiErr := gateway.performRequestForResponseBytes(request)
	if apiErr != nil {
		return
	}
	defer rawResponse.Body.Close()

	if rawResponse.StatusCode > 203 || strings.TrimSpace(string(bytes)) == "" {
		return
	}

	err := json.Unmarshal(bytes, &response)
	if err != nil {
		apiErr = errors.NewWithError(T("Invalid JSON response from server"), err)
		return
	}

	asyncResource := &AsyncResource{}
	err = json.Unmarshal(bytes, &asyncResource)
	if err != nil {
		apiErr = errors.NewWithError(T("Invalid async response from server"), err)
		return
	}

	jobUrl := asyncResource.Metadata.URL
	if jobUrl == "" {
		return
	}

	if !strings.Contains(jobUrl, "/jobs/") {
		return
	}

	apiErr = gateway.waitForJob(endpoint+jobUrl, request.HttpReq.Header.Get("Authorization"), timeout)

	return
}

func (gateway Gateway) Warnings() []string {
	return *gateway.warnings
}

func (gateway Gateway) waitForJob(jobUrl, accessToken string, timeout time.Duration) (err error) {
	startTime := gateway.Clock()
	for true {
		if gateway.Clock().Sub(startTime) > timeout && timeout != 0 {
			return errors.NewAsyncTimeoutError(jobUrl)
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
	if gateway.transport == nil {
		makeHttpTransport(&gateway)
	}

	httpClient := NewHttpClient(gateway.transport)

	dumpRequest(request)

	for i := 0; i < 3; i++ {
		response, err = httpClient.Do(request)
		if response == nil && err != nil {
			continue
		} else {
			break
		}
	}

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

func makeHttpTransport(gateway *Gateway) {
	gateway.transport = &http.Transport{
		Dial:            (&net.Dialer{Timeout: 5 * time.Second}).Dial,
		TLSClientConfig: NewTLSConfig(gateway.trustedCerts, gateway.config.IsSSLDisabled()),
		Proxy:           http.ProxyFromEnvironment,
	}
}

func (gateway *Gateway) SetTrustedCerts(certificates []tls.Certificate) {
	gateway.trustedCerts = certificates
	makeHttpTransport(gateway)
}
