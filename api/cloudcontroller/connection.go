package cloudcontroller

//go:generate counterfeiter . Connection

// Connection creates and executes http requests
type Connection interface {
	Make(passedRequest Request, passedResponse *Response) error
}
