package cfnetworking

// Connection creates and executes http requests
//
//counterfeiter:generate . Connection
type Connection interface {
	Make(request *Request, passedResponse *Response) error
}
