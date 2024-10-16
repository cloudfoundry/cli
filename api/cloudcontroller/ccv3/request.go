package ccv3

import (
	"io"
	"net/http"
	"strings"

	"code.cloudfoundry.org/cli/v8/api/cloudcontroller"
	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccv3/internal"
)

// requestOptions contains all the options to create an HTTP request.
type requestOptions struct {
	// URIParams are the list URI route parameters
	URIParams internal.Params

	// Query is a list of HTTP query parameters. Query will overwrite any
	// existing query string in the URI. If you want to preserve the query
	// string in URI make sure Query is nil.
	Query []Query

	// RequestName is the name of the request (see routes)
	RequestName string
	// Header contains custom headers to pass to the request.
	Header http.Header
	// Method is the HTTP method.
	Method string
	// URL is the request path.
	URL string
	// Body is the content of the request.
	Body io.ReadSeeker
}

// newHTTPRequest returns a constructed HTTP.Request with some defaults.
// Defaults are applied when Request options are not filled in.
func (requester *RealRequester) newHTTPRequest(passedRequest requestOptions) (*cloudcontroller.Request, error) {
	var request *http.Request
	var err error
	if passedRequest.URL != "" {
		request, err = http.NewRequest(
			passedRequest.Method,
			passedRequest.URL,
			passedRequest.Body,
		)
	} else {
		request, err = requester.router.CreateRequest(
			passedRequest.RequestName,
			map[string]string(passedRequest.URIParams),
			passedRequest.Body,
		)
	}
	if err != nil {
		return nil, err
	}

	if passedRequest.Query != nil {
		request.URL.RawQuery = FormatQueryParameters(passedRequest.Query).Encode()
	}

	request.Header = http.Header{}

	if passedRequest.Header != nil {
		request.Header = passedRequest.Header
	}

	if request.Header.Get("User-Agent") == "" {
		request.Header.Set("User-Agent", requester.userAgent)
	}

	if request.Header.Get("Accept") == "" {
		request.Header.Set("Accept", "application/json")
	}

	if !isDownloadDroplet(passedRequest.URL, passedRequest.RequestName) && request.Header.Get("Content-Type") == "" {
		request.Header.Set("Content-Type", "application/json")
	} else if isDownloadDroplet(passedRequest.URL, passedRequest.RequestName) && request.Header.Get("Content-Type") != "" {
		request.Header.Del("Content-Type")
	}

	return cloudcontroller.NewRequest(request, passedRequest.Body), nil
}

func isDownloadDroplet(URL string, requestName string) bool {
	return (strings.Contains(URL, "droplet") && strings.Contains(URL, "download")) || (requestName == internal.GetDropletBitsRequest)
}
