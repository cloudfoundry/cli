package cloudcontroller

import "net/http"

type Response struct {
	Result       interface{}
	RawResponse  []byte
	Warnings     []string
	HTTPResponse *http.Response
}

func (r *Response) reset() {
	r.RawResponse = []byte{}
	r.Warnings = []string{}
	r.HTTPResponse = nil
}
