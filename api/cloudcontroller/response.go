package cloudcontroller

import "net/http"

type Response struct {
	Result       interface{}
	RawResponse  []byte
	Warnings     []string
	HTTPResponse *http.Response
}
