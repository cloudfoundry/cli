package cloudcontroller

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// CloudControllerConnection represents a connection to the Cloud Controller
// server.
type CloudControllerConnection struct {
	HTTPClient *http.Client
	UserAgent  string
}

// Config is for configuring a CloudControllerConnection.
type Config struct {
	DialTimeout       time.Duration
	SkipSSLValidation bool
}

// NewConnection returns a new CloudControllerConnection with provided
// configuration.
func NewConnection(config Config) *CloudControllerConnection {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: config.SkipSSLValidation,
		},
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			KeepAlive: 30 * time.Second,
			Timeout:   config.DialTimeout,
		}).DialContext,
	}

	return &CloudControllerConnection{
		HTTPClient: &http.Client{Transport: tr},
	}
}

// Make performs the request and parses the response.
func (connection *CloudControllerConnection) Make(request *http.Request, passedResponse *Response) error {
	// In case this function is called from a retry, passedResponse may already
	// be populated with a previous response. We reset in case there's an HTTP
	// error and we don't repopulate it in populateResponse.
	passedResponse.reset()

	response, err := connection.HTTPClient.Do(request)
	if err != nil {
		return connection.processRequestErrors(request, err)
	}

	return connection.populateResponse(response, passedResponse)
}

func (connection *CloudControllerConnection) processRequestErrors(request *http.Request, err error) error {
	switch e := err.(type) {
	case *url.Error:
		switch urlErr := e.Err.(type) {
		case x509.UnknownAuthorityError:
			return UnverifiedServerError{
				URL: request.URL.String(),
			}
		case x509.HostnameError:
			return SSLValidationHostnameError{
				Message: urlErr.Error(),
			}
		default:
			return RequestError{Err: e}
		}
	default:
		return err
	}
}

func (connection *CloudControllerConnection) populateResponse(response *http.Response, passedResponse *Response) error {
	passedResponse.HTTPResponse = response

	// The cloud controller returns warnings with key "X-Cf-Warnings", and the
	// value is a comma seperated string.
	if rawWarnings := response.Header.Get("X-Cf-Warnings"); rawWarnings != "" {
		passedResponse.Warnings = []string{}
		for _, warning := range strings.Split(rawWarnings, ",") {
			warningTrimmed := strings.Trim(warning, " ")
			passedResponse.Warnings = append(passedResponse.Warnings, warningTrimmed)
		}
	}

	rawBytes, err := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		return err
	}

	passedResponse.RawResponse = rawBytes

	err = connection.handleStatusCodes(response, passedResponse)
	if err != nil {
		return err
	}

	if passedResponse.Result != nil {
		decoder := json.NewDecoder(bytes.NewBuffer(passedResponse.RawResponse))
		decoder.UseNumber()
		err = decoder.Decode(passedResponse.Result)
		if err != nil {
			return err
		}
	}

	return nil
}

func (*CloudControllerConnection) handleStatusCodes(response *http.Response, passedResponse *Response) error {
	if response.StatusCode >= 400 {
		return RawHTTPStatusError{
			StatusCode:  response.StatusCode,
			RawResponse: passedResponse.RawResponse,
		}
	}

	return nil
}
