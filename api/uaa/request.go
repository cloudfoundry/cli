package uaa

import (
	"io"
	"net/http"
	"net/url"

	"github.com/tedsuo/rata"
)

// RequestOptions contains all the options to create an HTTP Request.
type requestOptions struct {
	// Header is the set of request headers
	Header http.Header

	// Params are the list URI route parameters
	Params rata.Params

	// Query is a list of HTTP query parameters
	Query url.Values

	// RequestName is the name of the request (see routes)
	RequestName string

	// Body is the request body
	Body io.Reader
}

// newRequest returns a constructed http.Request with some defaults. The
// request will terminate the connection after it is sent (via a 'Connection:
// close' header).
func (client *Client) newRequest(passedRequest requestOptions) (*http.Request, error) {
	request, err := client.router.CreateRequest(
		passedRequest.RequestName,
		passedRequest.Params,
		passedRequest.Body,
	)
	if err != nil {
		return nil, err
	}

	request.URL.RawQuery = passedRequest.Query.Encode()

	if passedRequest.Header != nil {
		request.Header = passedRequest.Header
	} else {
		request.Header = http.Header{}
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Connection", "close")
	request.Header.Set("User-Agent", client.userAgent)

	return request, nil
}
