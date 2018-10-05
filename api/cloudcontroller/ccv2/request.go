package ccv2

import (
	"io"
	"net/http"
	"net/url"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
)

// Params represents URI parameters for a request.
type Params map[string]string

// requestOptions contains all the options to create an HTTP request.
type requestOptions struct {
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
func (client Client) newHTTPRequest(passedRequest requestOptions) (*cloudcontroller.Request, error) {
	var request *http.Request
	var err error
	if passedRequest.URI != "" {
		var (
			path *url.URL
			base *url.URL
		)

		path, err = url.Parse(passedRequest.URI)
		if err != nil {
			return nil, err
		}

		base, err = url.Parse(client.API())
		if err != nil {
			return nil, err
		}

		request, err = http.NewRequest(
			passedRequest.Method,
			base.ResolveReference(path).String(),
			passedRequest.Body,
		)
	} else {
		request, err = client.router.CreateRequest(
			passedRequest.RequestName,
			map[string]string(passedRequest.URIParams),
			passedRequest.Body,
		)
		if err == nil {
			request.URL.RawQuery = passedRequest.Query.Encode()
		}
	}
	if err != nil {
		return nil, err
	}

	request.Header = http.Header{}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", client.userAgent)

	if passedRequest.Body != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	// Make sure the body is the same as the one in the request
	return cloudcontroller.NewRequest(request, passedRequest.Body), nil
}
