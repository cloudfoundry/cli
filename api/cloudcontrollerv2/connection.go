package cloudcontrollerv2

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/tedsuo/rata"
)

type Request struct {
	RequestName string
	Params      rata.Params
	Query       url.Values

	URI    string
	Method string
}

type Response struct {
	Result   interface{}
	Warnings []string
}

type Connection struct {
	HTTPClient       *http.Client
	URL              string
	requestGenerator *rata.RequestGenerator
}

func NewConnection(APIURL string, skipSSLValidation bool) *Connection {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: skipSSLValidation,
		},
	}

	return &Connection{
		HTTPClient: &http.Client{Transport: tr},

		URL:              strings.TrimRight(APIURL, "/"),
		requestGenerator: rata.NewRequestGenerator(APIURL, Routes),
	}
}

func (connection *Connection) Make(passedRequest Request, passedResponse *Response) error {
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

func (connection *Connection) createHTTPRequest(passedRequest Request) (*http.Request, error) {
	var req *http.Request
	var err error
	if passedRequest.URI != "" {
		req, err = http.NewRequest(
			passedRequest.Method,
			fmt.Sprintf("%s%s", connection.URL, passedRequest.URI),
			&bytes.Buffer{},
		)
	} else {
		req, err = connection.requestGenerator.CreateRequest(
			passedRequest.RequestName,
			passedRequest.Params,
			&bytes.Buffer{},
		)
		req.URL.RawQuery = passedRequest.Query.Encode()
	}
	if err != nil {
		return nil, err
	}

	// for h, vs := range passedRequest.Header {
	// 	for _, v := range vs {
	// 		req.Header.Add(h, v)
	// 	}
	// }

	return req, nil
}

func (connection *Connection) processRequestErrors(err error) error {
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

func (connection *Connection) populateResponse(response *http.Response, passedResponse *Response) error {
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
		decoder := json.NewDecoder(response.Body)
		decoder.UseNumber()
		err = decoder.Decode(passedResponse.Result)
		if err != nil {
			return err
		}
	}

	return nil
}

func (*Connection) handleStatusCodes(response *http.Response) error {
	switch response.StatusCode {
	case http.StatusNotFound:
		var notFoundErr ResourceNotFoundError

		decoder := json.NewDecoder(response.Body)
		err := decoder.Decode(&notFoundErr)
		if err != nil {
			return err
		}

		return notFoundErr
	case http.StatusUnauthorized:
		return UnauthorizedError{}
	case http.StatusForbidden:
		return ForbiddenError{}
	}
	return nil
}
