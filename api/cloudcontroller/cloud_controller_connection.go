package cloudcontroller

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type CloudControllerConnection struct {
	HTTPClient *http.Client
}

func NewConnection(skipSSLValidation bool) *CloudControllerConnection {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: skipSSLValidation,
		},
		Proxy: http.ProxyFromEnvironment,
	}

	return &CloudControllerConnection{
		HTTPClient: &http.Client{Transport: tr},
	}
}

func (connection *CloudControllerConnection) Make(request *http.Request, passedResponse *Response) error {
	response, err := connection.HTTPClient.Do(request)
	if err != nil {
		return connection.processRequestErrors(request, err)
	}

	return connection.populateResponse(response, passedResponse)
}

func (connection *CloudControllerConnection) processRequestErrors(request *http.Request, err error) error {
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

func (connection *CloudControllerConnection) populateResponse(response *http.Response, passedResponse *Response) error {
	passedResponse.HTTPResponse = response

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
