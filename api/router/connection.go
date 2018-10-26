// Package router contains utilities to make call to the router API
package router

//go:generate counterfeiter . Connection

// Connection creates and executes http requests
type Connection interface {
	Make(request *Request, passedResponse *Response) error
}
