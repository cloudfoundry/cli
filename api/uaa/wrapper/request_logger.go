package wrapper

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"regexp"
	"sort"
	"time"

	"code.cloudfoundry.org/cli/api/uaa"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . RequestLoggerOutput

// RequestLoggerOutput is the interface for displaying logs
type RequestLoggerOutput interface {
	DisplayBody(body []byte) error
	DisplayJSONBody(body []byte) error
	DisplayHeader(name string, value string) error
	DisplayHost(name string) error
	DisplayRequestHeader(method string, uri string, httpProtocol string) error
	DisplayResponseHeader(httpProtocol string, status string) error
	DisplayType(name string, requestDate time.Time) error
	HandleInternalError(err error)
	Start() error
	Stop() error
}

// RequestLogger is the wrapper that logs requests to and responses from the
// UAA server
type RequestLogger struct {
	connection uaa.Connection
	output     RequestLoggerOutput
}

// NewRequestLogger returns a pointer to a RequestLogger wrapper
func NewRequestLogger(output RequestLoggerOutput) *RequestLogger {
	return &RequestLogger{
		output: output,
	}
}

// Make records the request and the response to UI
func (logger *RequestLogger) Make(request *http.Request, passedResponse *uaa.Response) error {
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

// Wrap sets the connection on the RequestLogger and returns itself
func (logger *RequestLogger) Wrap(innerconnection uaa.Connection) uaa.Connection {
	logger.connection = innerconnection
	return logger
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

	if request.Body != nil {
		rawRequestBody, err := ioutil.ReadAll(request.Body)
		defer request.Body.Close()
		if err != nil {
			return err
		}

		request.Body = ioutil.NopCloser(bytes.NewBuffer(rawRequestBody))
		if request.Header.Get("Content-Type") == "application/json" {
			err = logger.output.DisplayJSONBody(rawRequestBody)
		} else {
			err = logger.output.DisplayBody(rawRequestBody)
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func (logger *RequestLogger) displayResponse(passedResponse *uaa.Response) error {
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
	for key := range headers {
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
	redactedValue := "[PRIVATE DATA HIDDEN]"
	redactedKeys := []string{"Authorization", "Set-Cookie"}
	for _, redactedKey := range redactedKeys {
		if key == redactedKey {
			return redactedValue
		}
	}

	re := regexp.MustCompile(`([&?]code)=[A-Za-z0-9\-._~!$'()*+,;=:@/?]*`)
	if key == "Location" {
		value = re.ReplaceAllString(value, "$1="+redactedValue)
	}

	return value
}
