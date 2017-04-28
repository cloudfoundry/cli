// Package cloudcontroller contains shared utilies between the V2 and V3
// clients.
//
// These sets of packages are still under development/pre-pre-pre...alpha. Use
// at your own risk! Functionality and design may change without warning.
//
// Where are the clients?
//
// These clients live in ccv2 and ccv3 packages. Each of them only works with
// the V2 and V3 api respectively.
package cloudcontroller

//go:generate counterfeiter . Connection

// Connection creates and executes http requests
type Connection interface {
	Make(request *Request, passedResponse *Response) error
}
