package plugin

import (
	"bytes"
	"crypto/x509"
	"encoding/json"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"time"
	"errors"

	"code.cloudfoundry.org/cli/api/plugin/pluginerror"
	"code.cloudfoundry.org/cli/util"
)

// PluginConnection represents a connection to a plugin repo.
type PluginConnection struct {
	HTTPClient  *http.Client
	proxyReader ProxyReader // nolint
}

// NewConnection returns a new PluginConnection
func NewConnection(skipSSLValidation bool, dialTimeout time.Duration) *PluginConnection {
	tr := &http.Transport{
		TLSClientConfig: util.NewTLSConfig(nil, skipSSLValidation),
		Proxy:           http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			KeepAlive: 30 * time.Second,
			Timeout:   dialTimeout,
		}).DialContext,
	}

	return &PluginConnection{
		HTTPClient: &http.Client{Transport: tr},
	}
}

// Make performs the request and parses the response.
func (connection *PluginConnection) Make(request *http.Request, passedResponse *Response, proxyReader ProxyReader) error {
	// In case this function is called from a retry, passedResponse may already
	// be populated with a previous response. We reset in case there's an HTTP
	// error and we don't repopulate it in populateResponse.
	passedResponse.reset()

	response, err := connection.HTTPClient.Do(request)
	if err != nil {
		return connection.processRequestErrors(request, err)
	}

	body := response.Body
	if proxyReader != nil {
		proxyReader.Start(response.ContentLength)
		defer proxyReader.Finish()
		body = proxyReader.Wrap(response.Body)
	}

	return connection.populateResponse(response, passedResponse, body)
}

func (*PluginConnection) handleStatusCodes(response *http.Response, passedResponse *Response) error {
	if response.StatusCode >= 400 {
		return pluginerror.RawHTTPStatusError{
			Status:      response.Status,
			RawResponse: passedResponse.RawResponse,
		}
	}

	return nil
}

func (connection *PluginConnection) populateResponse(response *http.Response, passedResponse *Response, body io.ReadCloser) error {
	passedResponse.HTTPResponse = response

	rawBytes, err := ioutil.ReadAll(body)
	defer body.Close()
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

// processRequestError handles errors that occur while making the request.
func (connection *PluginConnection) processRequestErrors(request *http.Request, err error) error {
	switch e := err.(type) {
	case *url.Error:
		if errors.As(err, &x509.UnknownAuthorityError{}) {
			return pluginerror.UnverifiedServerError{
				URL: request.URL.String(),
			}
		}

		hostnameError := x509.HostnameError{}
		if errors.As(err, &hostnameError) {
			return pluginerror.SSLValidationHostnameError{
				Message: hostnameError.Error(),
			}
		}

		return pluginerror.RequestError{Err: e}

	default:
		return err
	}
}
