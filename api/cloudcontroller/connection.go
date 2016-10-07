package cloudcontroller

import (
	"net/http"
	"net/url"

	"github.com/tedsuo/rata"
)

type Request struct {
	Header      http.Header
	Params      rata.Params
	Query       url.Values
	RequestName string

	URI    string
	Method string
}

type Response struct {
	Result      interface{}
	RawResponse []byte
	Warnings    []string
}

//go:generate counterfeiter . Connection

// Connection creates and executes http requests
type Connection interface {
	Make(passedRequest Request, passedResponse *Response) error
}
