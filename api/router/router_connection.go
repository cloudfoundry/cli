package router

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

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/router/routererror"
)

// ConnectionConfig is for configuring the RouterConnection
type ConnectionConfig struct {
	DialTimeout       time.Duration
	SkipSSLValidation bool
}

// RouterConnection represents the connection to Router
type RouterConnection struct {
	HTTPClient *http.Client
}

// NewConnection returns a pointer to a new RouterConnection with the provided configuration
func NewConnection(config ConnectionConfig) *RouterConnection {
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

	return &RouterConnection{
		HTTPClient: &http.Client{Transport: tr},
	}
}

// Make performs the request and parses the response.
func (connection *RouterConnection) Make(request *Request, responseToPopulate *Response) error {
	// In case this function is called from a retry, passedResponse may already
	// be populated with a previous response. We reset in case there's an HTTP
	// error and we don't repopulate it in populateResponse.
	responseToPopulate.reset()

	httpResponse, err := connection.HTTPClient.Do(request.Request)
	if err != nil {
		// request could not be made, e.g., ssl handshake or tcp dial timeout
		// TODO: check on this
		return connection.processRequestErrors(request.Request, err)
	}

	return connection.populateResponse(httpResponse, responseToPopulate)
}

func (*RouterConnection) handleStatusCodes(httpResponse *http.Response, responseToPopulate *Response) error {
	if httpResponse.StatusCode >= 400 {
		var errorResponse routererror.ErrorResponse
		err := json.Unmarshal(responseToPopulate.RawResponse, &errorResponse)
		if err != nil {
			return routererror.RawHTTPStatusError{
				StatusCode:  httpResponse.StatusCode,
				RawResponse: responseToPopulate.RawResponse,
			}
		}
		errorResponse.StatusCode = httpResponse.StatusCode

		return errorResponse
	}

	return nil
}

func (connection *RouterConnection) populateResponse(httpResponse *http.Response, responseToPopulate *Response) error {
	responseToPopulate.HTTPResponse = httpResponse

	rawBytes, err := ioutil.ReadAll(httpResponse.Body)
	defer httpResponse.Body.Close()
	if err != nil {
		return err
	}
	responseToPopulate.RawResponse = rawBytes

	err = connection.handleStatusCodes(httpResponse, responseToPopulate)
	if err != nil {
		return err // TODO errConfig
	}

	if responseToPopulate.Result != nil {
		decoder := json.NewDecoder(bytes.NewBuffer(responseToPopulate.RawResponse))
		decoder.UseNumber()
		err = decoder.Decode(responseToPopulate.Result)
		if err != nil {
			return err
		}
	}

	return nil
}

func (*RouterConnection) processRequestErrors(request *http.Request, err error) error {
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
