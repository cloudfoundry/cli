package ccv3

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// requestOptions contains all the options to create an HTTP request.
type requestOptions struct {
	// Query is a list of HTTP query parameters
	Query url.Values

	URI    string
	Method string

	// Body is the request body
	Body io.Reader
}

// newHTTPRequest returns a constructed HTTP.Request with some defaults.
// Defaults are applied when Request fields are not filled in.
func (client CloudControllerClient) newHTTPRequest(passedRequest requestOptions) (*http.Request, error) {
	var request *http.Request
	var err error
	request, err = http.NewRequest(
		passedRequest.Method,
		fmt.Sprintf("%s%s", client.cloudControllerURL, passedRequest.URI),
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

	return request, nil
}
