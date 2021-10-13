package cfnetworking

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"time"

	"code.cloudfoundry.org/cfnetworking-cli-api/cfnetworking/networkerror"
)

// NetworkingConnection represents a connection to the Cloud Controller
// server.
type NetworkingConnection struct {
	HTTPClient *http.Client
	UserAgent  string
}

// Config is for configuring a NetworkingConnection.
type Config struct {
	DialTimeout       time.Duration
	SkipSSLValidation bool
}

// NewConnection returns a new NetworkingConnection with provided
// configuration.
func NewConnection(config Config) *NetworkingConnection {
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

	return &NetworkingConnection{
		HTTPClient: &http.Client{Transport: tr},
	}
}

// Make performs the request and parses the response.
func (connection *NetworkingConnection) Make(request *Request, passedResponse *Response) error {
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

func (*NetworkingConnection) processRequestErrors(request *http.Request, err error) error {
	switch e := err.(type) {
	case *url.Error:
		switch urlErr := e.Err.(type) {
		case x509.UnknownAuthorityError:
			return networkerror.UnverifiedServerError{
				URL: request.URL.String(),
			}
		case x509.HostnameError:
			return networkerror.SSLValidationHostnameError{
				Message: urlErr.Error(),
			}
		default:
			return networkerror.RequestError{Err: e}
		}
	default:
		return err
	}
}

func (connection *NetworkingConnection) populateResponse(response *http.Response, passedResponse *Response) error {
	passedResponse.HTTPResponse = response

	if resourceLocationURL := response.Header.Get("Location"); resourceLocationURL != "" {
		passedResponse.ResourceLocationURL = resourceLocationURL
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

func (*NetworkingConnection) handleStatusCodes(response *http.Response, passedResponse *Response) error {
	if response.StatusCode >= 400 {
		return networkerror.RawHTTPStatusError{
			StatusCode:  response.StatusCode,
			RawResponse: passedResponse.RawResponse,
			RequestIDs:  response.Header["X-Vcap-Request-Id"],
		}
	}

	return nil
}
