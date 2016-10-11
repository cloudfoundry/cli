package uaa

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/tedsuo/rata"
)

// UAAConnection represents the connection to UAA
type UAAConnection struct {
	HTTPClient       *http.Client
	URL              string
	requestGenerator *rata.RequestGenerator
}

// NewConnection returns a pointer to a new UAA Connection
func NewConnection(APIURL string, routes rata.Routes, skipSSLValidation bool) *UAAConnection {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: skipSSLValidation,
		},
	}

	return &UAAConnection{
		HTTPClient: &http.Client{Transport: tr},

		URL:              strings.TrimRight(APIURL, "/"),
		requestGenerator: rata.NewRequestGenerator(APIURL, routes),
	}
}

// Make takes a passedRequest, converts it into an HTTP request and then
// executes it. The response is then injected into passedResponse.
func (connection *UAAConnection) Make(passedRequest Request, passedResponse *Response) error {
	req, err := connection.createHTTPRequest(passedRequest)
	if err != nil {
		return err
	}

	response, err := connection.HTTPClient.Do(req)
	if err != nil {
		return connection.processRequestErrors(err)
	}

	defer response.Body.Close()

	return connection.populateResponse(response, passedResponse)
}

func (connection *UAAConnection) createHTTPRequest(passedRequest Request) (*http.Request, error) {
	request, err := connection.requestGenerator.CreateRequest(
		passedRequest.RequestName,
		passedRequest.Params,
		passedRequest.Body,
	)
	request.URL.RawQuery = passedRequest.Query.Encode()
	if err != nil {
		return nil, err
	}

	if passedRequest.Header != nil {
		request.Header = passedRequest.Header
	}

	// request.Header.Set("Connection", "close")
	// request.Header.Set("User-Agent", "go-cli "+cf.Version+" / "+runtime.GOOS)

	return request, nil
}

func (connection *UAAConnection) processRequestErrors(err error) error {
	switch e := err.(type) {
	case *url.Error:
		if _, ok := e.Err.(x509.UnknownAuthorityError); ok {
			return UnverifiedServerError{
				URL: connection.URL,
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
