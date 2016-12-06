package uaa

import "net/http"

// Response contains the result of a UAA request
type Response struct {
	// Result is the unserialized response of the UAA request
	Result interface{}

	// RawResponse is the raw bytes of the HTTP Response
	RawResponse []byte

	HTTPResponse *http.Response
}

func (r *Response) reset() {
	r.RawResponse = []byte{}
	r.HTTPResponse = nil
}
