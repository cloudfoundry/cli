package ccv3

import (
	"io"
	"net/http"
	"net/url"
)

// requestOptions contains all the options to create an HTTP request.
type requestOptions struct {
	// Query is a list of HTTP query parameters. Query will overwrite any
	// existing query string in the URI. If you want to preserve the query
	// string in URI make sure Query is nil.
	Query url.Values
	// request path
	URL string
	// HTTP Method
	Method string
	// request body
	Body io.Reader
}

// newHTTPRequest returns a constructed HTTP.Request with some defaults.
// Defaults are applied when Request options are not filled in.
func (client *Client) newHTTPRequest(passedRequest requestOptions) (*http.Request, error) {
	var request *http.Request
	var err error

	request, err = http.NewRequest(
		passedRequest.Method,
		passedRequest.URL,
		passedRequest.Body,
	)
	if err != nil {
		return nil, err
	}

	if passedRequest.Query != nil {
		request.URL.RawQuery = passedRequest.Query.Encode()
	}

	request.Header = http.Header{}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", client.userAgent)

	return request, nil
}
