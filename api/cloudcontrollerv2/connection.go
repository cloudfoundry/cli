package cloudcontrollerv2

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/tedsuo/rata"
)

type Connection struct {
	HTTPClient       *http.Client
	URL              string
	requestGenerator *rata.RequestGenerator
}

type Request struct {
	RequestName string
	Params      rata.Params
	// Query       url.Values
	// Body        *bytes.Buffer
}

type Response struct {
	Result   interface{}
	Warnings []string
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
	body := connection.getBody(passedRequest)

	req, err := connection.requestGenerator.CreateRequest(
		passedRequest.RequestName,
		passedRequest.Params,
		body,
	)
	if err != nil {
		return nil, err
	}

	// req.URL.RawQuery = passedRequest.Query.Encode()

	// for h, vs := range passedRequest.Header {
	// 	for _, v := range vs {
	// 		req.Header.Add(h, v)
	// 	}
	// }

	return req, nil
}

func (connection *Connection) getBody(passedRequest Request) *bytes.Buffer {
	// if passedRequest.Body != nil {
	// 	if _, ok := passedRequest.Header["Content-Type"]; !ok {
	// 		panic("You must pass a 'Content-Type' Header with a body")
	// 	}
	// 	return passedRequest.Body
	// }

	return &bytes.Buffer{}
}

func (_ *Connection) processRequestErrors(err error) error {
	switch e := err.(type) {
	case *url.Error:
		if _, ok := e.Err.(x509.UnknownAuthorityError); ok {
			return UnverifiedServerError{}
		}
		return RequestError(e)
	default:
		return err
	}
}

func (connection *Connection) populateResponse(response *http.Response, passedResponse *Response) error {
	err := connection.handleStatusCodes(response)
	if err != nil {
		return err
	}

	if rawWarnings := response.Header.Get("X-Cf-Warnings"); rawWarnings != "" {
		passedResponse.Warnings = []string{}
		for _, warning := range strings.Split(rawWarnings, ",") {
			warningTrimmed := strings.Trim(warning, " ")
			passedResponse.Warnings = append(passedResponse.Warnings, warningTrimmed)
		}
	}

	// if response.StatusCode < 200 || response.StatusCode >= 300 {
	// 	body, _ := ioutil.ReadAll(response.Body)

	// 	return UnexpectedResponseError{
	// 		StatusCode: response.StatusCode,
	// 		Status:     response.Status,
	// 		Body:       string(body),
	// 	}
	// }

	// if passedResponse == nil {
	// 	return nil
	// }

	// switch response.StatusCode {
	// case http.StatusNoContent:
	// 	return nil
	// case http.StatusCreated:
	// 	passedResponse.Created = true
	// }

	// if passedResponse.Headers != nil {
	// 	for k, v := range response.Header {
	// 		(*passedResponse.Headers)[k] = v
	// 	}
	// }

	// if returnResponseBody {
	// 	passedResponse.Result = response.Body
	// 	return nil
	// }

	// if passedResponse.Result == nil {
	// 	return nil
	// }

	err = json.NewDecoder(response.Body).Decode(passedResponse.Result)
	if err != nil {
		return err
	}

	return nil
}

func (_ *Connection) handleStatusCodes(response *http.Response) error {
	switch response.StatusCode {
	case http.StatusNotFound:
		return ResourceNotFoundError{}
	case http.StatusUnauthorized:
		return UnauthorizedError{}
	case http.StatusForbidden:
		return ForbiddenError{}
	}
	return nil
}
