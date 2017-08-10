package cfnetworking

//go:generate counterfeiter . Connection

// Connection creates and executes http requests
type Connection interface {
	Make(request *Request, passedResponse *Response) error
}
