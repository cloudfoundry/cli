package cloudcontroller

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/tedsuo/rata"
)

type CloudControllerConnection struct {
	HTTPClient       *http.Client
	URL              string
	requestGenerator *rata.RequestGenerator
}

func NewConnection(APIURL string, routes rata.Routes, skipSSLValidation bool) *CloudControllerConnection {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: skipSSLValidation,
		},
	}

	return &CloudControllerConnection{
		HTTPClient: &http.Client{Transport: tr},

		URL:              strings.TrimRight(APIURL, "/"),
		requestGenerator: rata.NewRequestGenerator(APIURL, routes),
	}
}

func (connection *CloudControllerConnection) Make(passedRequest Request, passedResponse *Response) error {
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

func (connection *CloudControllerConnection) createHTTPRequest(passedRequest Request) (*http.Request, error) {
	var request *http.Request
	var err error
	if passedRequest.URI != "" {
		request, err = http.NewRequest(
			passedRequest.Method,
			fmt.Sprintf("%s%s", connection.URL, passedRequest.URI),
			&bytes.Buffer{},
		)
	} else {
		request, err = connection.requestGenerator.CreateRequest(
			passedRequest.RequestName,
			passedRequest.Params,
			&bytes.Buffer{},
		)
		if err == nil {
			request.URL.RawQuery = passedRequest.Query.Encode()
		}
	}
	if err != nil {
		return nil, err
	}

	if passedRequest.Header != nil {
		request.Header = passedRequest.Header
	}

	request.Header.Set("accept", "application/json")
	request.Header.Set("content-type", "application/json")

	// request.Header.Set("Connection", "close")
	// request.Header.Set("User-Agent", "go-cli "+cf.Version+" / "+runtime.GOOS)

	return request, nil
}

func (connection *CloudControllerConnection) processRequestErrors(err error) error {
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

func (connection *CloudControllerConnection) populateResponse(response *http.Response, passedResponse *Response) error {
	if rawWarnings := response.Header.Get("X-Cf-Warnings"); rawWarnings != "" {
		passedResponse.Warnings = []string{}
		for _, warning := range strings.Split(rawWarnings, ",") {
			warningTrimmed := strings.Trim(warning, " ")
			passedResponse.Warnings = append(passedResponse.Warnings, warningTrimmed)
		}
	}

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

func (*CloudControllerConnection) handleStatusCodes(response *http.Response) error {
	if response.StatusCode >= 400 {
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return err
		}

		return RawCCError{
			StatusCode:  response.StatusCode,
			RawResponse: body,
		}
	}

	return nil
}
