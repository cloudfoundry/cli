package wrapper

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"sort"
	"time"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
)

//go:generate counterfeiter . RequestLoggerOutput

type RequestLoggerOutput interface {
	DisplayBody(body []byte)
	DisplayHeader(name string, value string)
	DisplayHost(name string)
	DisplayRequestHeader(method string, uri string, httpProtocol string)
	DisplayResponseHeader(httpProtocol string, status string)
	DisplayType(name string, requestDate time.Time)
	HandleInternalError(err error)
	Start() error
	Stop() error
}

type RequestLogger struct {
	connection cloudcontroller.Connection
	output     RequestLoggerOutput
}

// NewRequestLogger returns a pointer to a RequestLogger wrapper
func NewRequestLogger(output RequestLoggerOutput) *RequestLogger {
	return &RequestLogger{
		output: output,
	}
}

// Wrap sets the connection on the RequestLogger and returns itself
func (logger *RequestLogger) Wrap(innerconnection cloudcontroller.Connection) cloudcontroller.Connection {
	logger.connection = innerconnection
	return logger
}

// Make records the request and the response to ui
func (logger *RequestLogger) Make(request *http.Request, passedResponse *cloudcontroller.Response) error {
	err := logger.displayRequest(request)
	if err != nil {
		logger.output.HandleInternalError(err)
	}

	err = logger.connection.Make(request, passedResponse)

	if passedResponse.HTTPResponse != nil {
		displayErr := logger.displayResponse(passedResponse)
		if err != nil {
			logger.output.HandleInternalError(displayErr)
		}
	}

	return err
}

func (logger *RequestLogger) displayRequest(request *http.Request) error {
	err := logger.output.Start()
	if err != nil {
		return err
	}
	defer logger.output.Stop()

	logger.output.DisplayType("REQUEST", time.Now())
	logger.output.DisplayRequestHeader(request.Method, request.URL.Path, request.Proto)
	logger.output.DisplayHost(request.URL.Host)
	logger.displaySortedHeaders(request.Header)

	if request.Body != nil {
		rawRequestBody, err := ioutil.ReadAll(request.Body)
		defer request.Body.Close()
		if err != nil {
			return err
		}
		logger.output.DisplayBody(rawRequestBody)
		request.Body = ioutil.NopCloser(bytes.NewBuffer(rawRequestBody))
	}

	return nil
}

func (logger *RequestLogger) displayResponse(passedResponse *cloudcontroller.Response) error {
	err := logger.output.Start()
	if err != nil {
		return err
	}
	defer logger.output.Stop()

	logger.output.DisplayType("RESPONSE", time.Now())
	logger.output.DisplayResponseHeader(passedResponse.HTTPResponse.Proto, passedResponse.HTTPResponse.Status)
	logger.displaySortedHeaders(passedResponse.HTTPResponse.Header)
	logger.output.DisplayBody(passedResponse.RawResponse)
	return nil
}

func (logger *RequestLogger) displaySortedHeaders(headers http.Header) {
	keys := []string{}
	for key, _ := range headers {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		for _, value := range headers[key] {
			logger.output.DisplayHeader(key, value)
		}
	}
}
