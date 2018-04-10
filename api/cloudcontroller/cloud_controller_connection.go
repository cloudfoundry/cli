package cloudcontroller

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
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
func (connection *CloudControllerConnection) Make(request *Request, passedResponse *Response) error {
	// In case this function is called from a retry, passedResponse may already
	// be populated with a previous response. We reset in case there's an HTTP
	// error and we don't repopulate it in populateResponse.
	passedResponse.reset()

	response, err := connection.HTTPClient.Do(request.Request)
	if err != nil {
		return connection.processRequestErrors(request.Request, err)
	}

	return connection.populateResponse(response, passedResponse)
}

func (*CloudControllerConnection) processRequestErrors(request *http.Request, err error) error {
	switch e := err.(type) {
	case *url.Error:
		switch urlErr := e.Err.(type) {
		case x509.UnknownAuthorityError:
			return ccerror.UnverifiedServerError{
				URL: request.URL.String(),
			}
		case x509.HostnameError:
			return ccerror.SSLValidationHostnameError{
				Message: urlErr.Error(),
			}
		default:
			return ccerror.RequestError{Err: e}
		}
	default:
		return err
	}
}

func (connection *CloudControllerConnection) populateResponse(response *http.Response, passedResponse *Response) error {
	passedResponse.HTTPResponse = response

	warnings, err := connection.handleWarnings(response)
	if err != nil {
		return err
	}
	passedResponse.Warnings = warnings

	if resourceLocationURL := response.Header.Get("Location"); resourceLocationURL != "" {
		passedResponse.ResourceLocationURL = resourceLocationURL
	}

	err = connection.handleStatusCodes(response, passedResponse)
	if err != nil {
		return err
	}

	// TODO: only unmarshal on 'application/json', skip otherwise - Fixing this
	// todo will require changing ALL the API tests to include the content-type
	// in their tests.
	if passedResponse.Result != nil {
		err = DecodeJSON(passedResponse.RawResponse, passedResponse.Result)
		if err != nil {
			return err
		}
	}

	return nil
}

// handleWarnings looks for the "X-Cf-Warnings" header in the cloud controller
// response and URI decodes them. The value can contain multiple warnings that
// are comma separated.
func (*CloudControllerConnection) handleWarnings(response *http.Response) ([]string, error) {
	rawWarnings := response.Header.Get("X-Cf-Warnings")
	rawWarnings, err := url.QueryUnescape(rawWarnings)
	if err != nil {
		return nil, err
	}

	var warnings []string
	if rawWarnings != "" {
		for _, warning := range strings.Split(rawWarnings, ",") {
			warningTrimmed := strings.Trim(warning, " ")
			warnings = append(warnings, warningTrimmed)
		}
	}

	return warnings, nil
}

func (*CloudControllerConnection) handleStatusCodes(response *http.Response, passedResponse *Response) error {
	if response.StatusCode == http.StatusNoContent {
		passedResponse.RawResponse = []byte("{}")
	} else {
		rawBytes, err := ioutil.ReadAll(response.Body)
		defer response.Body.Close()
		if err != nil {
			return err
		}

		passedResponse.RawResponse = rawBytes
	}

	if response.StatusCode >= 400 {
		return ccerror.RawHTTPStatusError{
			StatusCode:  response.StatusCode,
			RawResponse: passedResponse.RawResponse,
			RequestIDs:  response.Header["X-Vcap-Request-Id"],
		}
	}

	return nil
}
