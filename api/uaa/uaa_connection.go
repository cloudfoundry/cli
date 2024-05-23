package uaa

import (
	"bytes"
	"crypto/x509"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"code.cloudfoundry.org/cli/util"
)

// UAAConnection represents the connection to UAA
type UAAConnection struct {
	HTTPClient *http.Client
}

// NewConnection returns a pointer to a new UAA Connection
func NewConnection(skipSSLValidation bool, disableKeepAlives bool, dialTimeout time.Duration) *UAAConnection {
	tr := &http.Transport{
		DialContext: (&net.Dialer{
			KeepAlive: 30 * time.Second,
			Timeout:   dialTimeout,
		}).DialContext,
		DisableKeepAlives: disableKeepAlives,
		Proxy:             http.ProxyFromEnvironment,
		TLSClientConfig:   util.NewTLSConfig(nil, skipSSLValidation),
	}

	return &UAAConnection{
		HTTPClient: &http.Client{
			Transport: tr,
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
				// This prevents redirects. When making a request to /oauth/authorize,
				// the client should not follow redirects in order to obtain the ssh
				// passcode.
				return http.ErrUseLastResponse
			},
		},
	}
}

// Make takes a passedRequest, converts it into an HTTP request and then
// executes it. The response is then injected into passedResponse.
func (connection *UAAConnection) Make(request *http.Request, passedResponse *Response) error {
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

func (*UAAConnection) handleStatusCodes(response *http.Response, passedResponse *Response) error {
	if response.StatusCode >= 400 {
		return RawHTTPStatusError{
			StatusCode:  response.StatusCode,
			RawResponse: passedResponse.RawResponse,
		}
	}

	return nil
}

func (connection *UAAConnection) populateResponse(response *http.Response, passedResponse *Response) error {
	passedResponse.HTTPResponse = response

	rawBytes, err := io.ReadAll(response.Body)
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

func (connection *UAAConnection) processRequestErrors(request *http.Request, err error) error {
	switch e := err.(type) {
	case *url.Error:
		if errors.As(err, &x509.UnknownAuthorityError{}) {
			return UnverifiedServerError{
				URL: request.URL.String(),
			}
		}
		return RequestError{Err: e}
	default:
		return err
	}
}
