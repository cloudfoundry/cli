package errors

import "fmt"

type HttpError interface {
	Error
	StatusCode() int
	Headers() string
	Body() string
}

type httpError struct {
	statusCode  int
	headers     string
	body        string
	code        string
	description string
}

type HttpNotFoundError struct {
	*httpError
}

func NewHttpError(statusCode int, header string, body string, code string, description string) HttpError {
	err := httpError{
		statusCode:  statusCode,
		headers:     header,
		body:        body,
		code:        code,
		description: description,
	}
	switch statusCode {
	case 404:
		return HttpNotFoundError{&err}
	default:
		return &err
	}
}

func (err *httpError) StatusCode() int {
	return err.statusCode
}

func (err *httpError) Headers() string {
	return err.headers
}

func (err *httpError) Body() string {
	return err.body
}

func (err *httpError) Error() string {
	return fmt.Sprintf(
		"Server error, status code: %d, error code: %s, message: %s",
		err.statusCode,
		err.code,
		err.description,
	)
}

func (err *httpError) ErrorCode() string {
	return err.code
}
