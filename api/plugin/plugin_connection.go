package plugin

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"time"

	"code.cloudfoundry.org/cli/api/plugin/pluginerror"
)

// PluginConnection represents a connection to a plugin repo.
type PluginConnection struct {
	HTTPClient  *http.Client
	proxyReader ProxyReader
}

// NewConnection returns a new PluginConnection
func NewConnection(skipSSLValidation bool, dialTimeout time.Duration) *PluginConnection {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: skipSSLValidation,
		},
		Proxy: http.ProxyFromEnvironment,
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

// processRequestError handles errors that occur while making the request.
func (connection *PluginConnection) processRequestErrors(request *http.Request, err error) error {
	switch e := err.(type) {
	case *url.Error:
		switch urlErr := e.Err.(type) {
		case x509.UnknownAuthorityError:
			return pluginerror.UnverifiedServerError{
				URL: request.URL.String(),
			}
		case x509.HostnameError:
			return pluginerror.SSLValidationHostnameError{
				Message: urlErr.Error(),
			}
		default:
			return pluginerror.RequestError{Err: e}
		}
	default:
		return err
	}
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

func (*PluginConnection) handleStatusCodes(response *http.Response, passedResponse *Response) error {
	if response.StatusCode >= 400 {
		return pluginerror.RawHTTPStatusError{
			Status:      response.Status,
			RawResponse: passedResponse.RawResponse,
		}
	}

	return nil
}
