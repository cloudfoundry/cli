package uaa

import (
	"io"
	"net/http"
	"net/url"

	"github.com/tedsuo/rata"
)

// Request contains all the elements of a UAA request
type Request struct {
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

// NewRequest contains a constructed UAA request with some defaults. There is
// an optional body that can be passed, however only the first body is used.
func NewRequest(requestName string, params rata.Params, header http.Header, query url.Values, body ...io.Reader) Request {
	if header == nil {
		header = http.Header{}
	}
	header.Set("Accept", "application/json")

	request := Request{
		RequestName: requestName,
		Params:      params,
		Header:      header,
		Query:       query,
	}

	if len(body) == 1 {
		request.Body = body[0]
	}

	return request
}
