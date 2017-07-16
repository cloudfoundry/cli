package plugin

import "net/http"

//go:generate counterfeiter . Connection

// Connection creates and executes http requests
type Connection interface {
	Make(request *http.Request, passedResponse *Response, proxyReader ProxyReader) error
}
