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
	DisplayType(name string, requestDate time.Time)
	DisplayHost(name string)
	DisplayRequest(method string, uri string, httpProtocol string)
	DisplayHeader(name string, value string)
	DisplayBody(body []byte)
	DisplayResponseHeader(httpProtocol string, status string)
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
	logger.output.DisplayType("REQUEST", time.Now())
	logger.output.DisplayRequest(request.Method, request.URL.Path, request.Proto)
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

	err := logger.connection.Make(request, passedResponse)

	if passedResponse.HTTPResponse != nil {
		logger.output.DisplayType("RESPONSE", time.Now())
		logger.output.DisplayResponseHeader(passedResponse.HTTPResponse.Proto, passedResponse.HTTPResponse.Status)
		logger.displaySortedHeaders(passedResponse.HTTPResponse.Header)
		logger.output.DisplayBody(passedResponse.RawResponse)
	}

	return err
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
