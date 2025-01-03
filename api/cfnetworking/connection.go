package cfnetworking

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Connection

// Connection creates and executes http requests
type Connection interface {
    Make(request *Request, passedResponse *Response) error
}
