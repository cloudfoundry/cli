package cloudcontroller

import (
	"io"
	"net/http"
	"net/url"
)

// Request contains all the elements of a Cloud Controller request
type Request struct { // TODO fixme
	// Header is the set of request headers
	Header http.Header

	// URIParams are the list URI route parameters
	URIParams map[string]string

	// Query is a list of HTTP query parameters
	Query url.Values

	// RequestName is the name of the request (see routes)
	RequestName string

	URI    string
	Method string

	// Body is the request body
	Body io.Reader
}

// NewRequest contains a constructed UAA request with some defaults. There is
// an optional body that can be passed, however only the first body is used.
func NewRequest(requestName string, URIParams map[string]string, header http.Header, query url.Values, body ...io.Reader) Request {
	header = processHeader(header)

	request := Request{
		RequestName: requestName,
		URIParams:   URIParams,
		Header:      header,
		Query:       query,
	}

	if len(body) == 1 {
		request.Body = body[0]
	}

	return request
}

func NewRequestFromURI(uri string, method string, header http.Header) Request {
	header = processHeader(header)

	return Request{
		URI:    uri,
		Method: method,
		Header: header,
	}
}

func processHeader(header http.Header) http.Header {
	if header == nil {
		header = http.Header{}
	}

	header.Set("Accept", "application/json")
	header.Set("content-type", "application/json")

	// request.Header.Set("User-Agent", "go-cli "+cf.Version+" / "+runtime.GOOS)

	return header
}
