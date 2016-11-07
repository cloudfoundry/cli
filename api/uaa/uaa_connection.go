package uaa

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
)

// UAAConnection represents the connection to UAA
type UAAConnection struct {
	HTTPClient *http.Client
}

// NewConnection returns a pointer to a new UAA Connection
func NewConnection(skipSSLValidation bool, dialTimeout time.Duration) *UAAConnection {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: skipSSLValidation,
		},
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout: dialTimeout,
		}).DialContext,
	}

	return &UAAConnection{
		HTTPClient: &http.Client{Transport: tr},
	}
}

// Make takes a passedRequest, converts it into an HTTP request and then
// executes it. The response is then injected into passedResponse.
func (connection *UAAConnection) Make(request *http.Request, passedResponse *Response) error {
	response, err := connection.HTTPClient.Do(request)
	if err != nil {
		return connection.processRequestErrors(request, err)
	}

	defer response.Body.Close()

	return connection.populateResponse(response, passedResponse)
}

func (connection *UAAConnection) processRequestErrors(request *http.Request, err error) error {
	switch e := err.(type) {
	case *url.Error:
		if _, ok := e.Err.(x509.UnknownAuthorityError); ok {
			return UnverifiedServerError{
				URL: request.URL.String(),
			}
		}
		return RequestError{Err: e}
	default:
		return err
	}
}

func (connection *UAAConnection) populateResponse(response *http.Response, passedResponse *Response) error {
	err := connection.handleStatusCodes(response)
	if err != nil {
		return err
	}

	if passedResponse.Result != nil {
		rawBytes, _ := ioutil.ReadAll(response.Body)
		passedResponse.RawResponse = rawBytes

		decoder := json.NewDecoder(bytes.NewBuffer(rawBytes))
		decoder.UseNumber()
		err = decoder.Decode(passedResponse.Result)
		if err != nil {
			return err
		}
	}

	return nil
}

func (*UAAConnection) handleStatusCodes(response *http.Response) error {
	if response.StatusCode >= 400 {
		var uaaErr Error
		decoder := json.NewDecoder(response.Body)
		err := decoder.Decode(&uaaErr)
		if err != nil {
			return err
		}

		return uaaErr
	}

	return nil
}
