package wrapper

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
	"time"

	"code.cloudfoundry.org/cfnetworking-cli-api/cfnetworking"
)

//go:generate counterfeiter . RequestLoggerOutput

// RequestLoggerOutput is the interface for displaying logs
type RequestLoggerOutput interface {
	DisplayHeader(name string, value string) error
	DisplayHost(name string) error
	DisplayJSONBody(body []byte) error
	DisplayMessage(msg string) error
	DisplayRequestHeader(method string, uri string, httpProtocol string) error
	DisplayResponseHeader(httpProtocol string, status string) error
	DisplayType(name string, requestDate time.Time) error
	HandleInternalError(err error)
	Start() error
	Stop() error
}

// RequestLogger is the wrapper that logs requests to and responses from the
// Cloud Controller server
type RequestLogger struct {
	connection cfnetworking.Connection
	output     RequestLoggerOutput
}

// NewRequestLogger returns a pointer to a RequestLogger wrapper
func NewRequestLogger(output RequestLoggerOutput) *RequestLogger {
	return &RequestLogger{
		output: output,
	}
}

// Wrap sets the connection on the RequestLogger and returns itself
func (logger *RequestLogger) Wrap(innerconnection cfnetworking.Connection) cfnetworking.Connection {
	logger.connection = innerconnection
	return logger
}

// Make records the request and the response to UI
func (logger *RequestLogger) Make(request *cfnetworking.Request, passedResponse *cfnetworking.Response) error {
	err := logger.displayRequest(request)
	if err != nil {
		logger.output.HandleInternalError(err)
	}

	err = logger.connection.Make(request, passedResponse)

	if passedResponse.HTTPResponse != nil {
		displayErr := logger.displayResponse(passedResponse)
		if displayErr != nil {
			logger.output.HandleInternalError(displayErr)
		}
	}

	return err
}

func (logger *RequestLogger) displayRequest(request *cfnetworking.Request) error {
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

	contentType := request.Header.Get("Content-Type")
	if request.Body != nil {
		if strings.Contains(contentType, "json") {
			rawRequestBody, err := ioutil.ReadAll(request.Body)
			if err != nil {
				return err
			}

			defer request.ResetBody()

			return logger.output.DisplayJSONBody(rawRequestBody)
		} else if contentType != "" {
			return logger.output.DisplayMessage(fmt.Sprintf("[%s Content Hidden]", strings.Split(contentType, ";")[0]))
		}
	}
	return nil
}

func (logger *RequestLogger) displayResponse(passedResponse *cfnetworking.Response) error {
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
