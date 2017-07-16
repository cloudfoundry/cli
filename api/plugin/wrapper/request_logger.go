package wrapper

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"sort"
	"time"

	"code.cloudfoundry.org/cli/api/plugin"
)

//go:generate counterfeiter . RequestLoggerOutput

// RequestLoggerOutput is the interface for displaying logs
type RequestLoggerOutput interface {
	DisplayDump(dump string) error
	DisplayHeader(name string, value string) error
	DisplayHost(name string) error
	DisplayJSONBody(body []byte) error
	DisplayRequestHeader(method string, uri string, httpProtocol string) error
	DisplayResponseHeader(httpProtocol string, status string) error
	DisplayType(name string, requestDate time.Time) error
	HandleInternalError(err error)
	Start() error
	Stop() error
}

// RequestLogger is the wrapper that logs requests to and responses from
// a plugin repository
type RequestLogger struct {
	connection plugin.Connection
	output     RequestLoggerOutput
}

// NewRequestLogger returns a pointer to a RequestLogger wrapper
func NewRequestLogger(output RequestLoggerOutput) *RequestLogger {
	return &RequestLogger{
		output: output,
	}
}

// Wrap sets the connection on the RequestLogger and returns itself
func (logger *RequestLogger) Wrap(innerconnection plugin.Connection) plugin.Connection {
	logger.connection = innerconnection
	return logger
}

// Make records the request and the response to UI
func (logger *RequestLogger) Make(request *http.Request, passedResponse *plugin.Response, proxyReader plugin.ProxyReader) error {
	err := logger.displayRequest(request)
	if err != nil {
		logger.output.HandleInternalError(err)
	}

	err = logger.connection.Make(request, passedResponse, proxyReader)

	if passedResponse.HTTPResponse != nil {
		displayErr := logger.displayResponse(passedResponse)
		if displayErr != nil {
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

	err = logger.output.DisplayType("REQUEST", time.Now())
	if err != nil {
		return err
	}
	err = logger.output.DisplayRequestHeader(request.Method, request.URL.RequestURI(), request.Proto)
	if err != nil {
		return err
	}
	err = logger.output.DisplayHost(request.URL.Host)
	if err != nil {
		return err
	}
	err = logger.displaySortedHeaders(request.Header)
	if err != nil {
		return err
	}

	if request.Body != nil && request.Header.Get("Content-Type") == "application/json" {
		rawRequestBody, err := ioutil.ReadAll(request.Body)
		defer request.Body.Close()
		if err != nil {
			return err
		}

		request.Body = ioutil.NopCloser(bytes.NewBuffer(rawRequestBody))
		err = logger.output.DisplayJSONBody(rawRequestBody)
		if err != nil {
			return err
		}
	}

	return nil
}

func (logger *RequestLogger) displayResponse(passedResponse *plugin.Response) error {
	err := logger.output.Start()
	if err != nil {
		return err
	}
	defer logger.output.Stop()

	err = logger.output.DisplayType("RESPONSE", time.Now())
	if err != nil {
		return err
	}
	err = logger.output.DisplayResponseHeader(passedResponse.HTTPResponse.Proto, passedResponse.HTTPResponse.Status)
	if err != nil {
		return err
	}
	err = logger.displaySortedHeaders(passedResponse.HTTPResponse.Header)
	if err != nil {
		return err
	}

	if passedResponse.HTTPResponse.Body == nil {
		return nil
	}
	if passedResponse.HTTPResponse.Header.Get("Content-Type") != "application/json" {
		return logger.output.DisplayDump("[NON-JSON BODY CONTENT HIDDEN]")
	}

	return logger.output.DisplayJSONBody(passedResponse.RawResponse)
}

func (logger *RequestLogger) displaySortedHeaders(headers http.Header) error {
	keys := []string{}
	for key, _ := range headers {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		for _, value := range headers[key] {
			err := logger.output.DisplayHeader(key, redactHeaders(key, value))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func redactHeaders(key string, value string) string {
	if key == "Authorization" {
		return "[PRIVATE DATA HIDDEN]"
	}
	return value
}
