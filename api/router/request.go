package router

import (
	"io"
	"net/http"
	"net/url"

	"github.com/tedsuo/rata"
)

// Request represents the request of the router
type Request struct {
	*http.Request

	body io.ReadSeeker
}

func (r *Request) ResetBody() error {
	if r.body == nil {
		return nil
	}

	_, err := r.body.Seek(0, 0)
	return err
}

// Params represents URI parameters for a request.
type Params map[string]string

// requestOptions contains all the options to create an HTTP request.
type requestOptions struct {
	// Header is the set of request headers
	Header http.Header

	// Body is the request body
	Body io.ReadSeeker

	// Method is the HTTP method of the request.
	Method string

	// Query is a list of HTTP query parameters
	Query url.Values

	// RequestName is the name of the request (see routes)
	RequestName string

	// URI is the URI of the request.
	URI string

	// URIParams are the list URI route parameters
	URIParams Params
}

// newHTTPRequest returns a constructed HTTP.Request with some defaults.
// Defaults are applied when Request fields are not filled in.
func (client Client) newHTTPRequest(passedRequest requestOptions) (*Request, error) {
	request, err := client.router.CreateRequest(
		passedRequest.RequestName,
		rata.Params(passedRequest.URIParams),
		passedRequest.Body,
	)

	if err != nil {
		return nil, err
	}

	if passedRequest.Query != nil {
		request.URL.RawQuery = passedRequest.Query.Encode()
	}

	if passedRequest.Header != nil {
		request.Header = passedRequest.Header
	} else {
		request.Header = http.Header{}
	}

	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Connection", "close")
	request.Header.Set("User-Agent", client.userAgent)

	return &Request{Request: request, body: passedRequest.Body}, nil
}

func NewRequest(request *http.Request, body io.ReadSeeker) *Request {
	return &Request{
		Request: request,
		body:    body,
	}
}
