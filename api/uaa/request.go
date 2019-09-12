package uaa

import (
	"io"
	"net/http"
	"net/url"

	"code.cloudfoundry.org/cli/api/uaa/internal"
)

// RequestOptions contains all the options to create an HTTP Request.
type requestOptions struct {
	// Header is the set of request headers
	Header http.Header

	// URIParams are the list URI route parameters
	URIParams internal.Params

	// Query is a list of HTTP query parameters
	Query url.Values

	// RequestName is the name of the request (see routes)
	RequestName string

	// Method is the HTTP method.
	Method string
	// URL is the request path.
	URL string
	// Body is the request body
	Body io.Reader
}

// newRequest returns a constructed http.Request with some defaults. The
// request will terminate the connection after it is sent (via a 'Connection:
// close' header).
func (client *Client) newRequest(passedRequest requestOptions) (*http.Request, error) {
	var request *http.Request
	var err error

	if passedRequest.URL != "" {
		request, err = http.NewRequest(
			passedRequest.Method,
			passedRequest.URL,
			passedRequest.Body,
		)
	} else {
		request, err = client.router.CreateRequest(
			passedRequest.RequestName,
			passedRequest.URIParams,
			passedRequest.Body,
		)
	}
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
	request.Header.Set("Connection", "close")
	request.Header.Set("User-Agent", client.userAgent)

	return request, nil
}
