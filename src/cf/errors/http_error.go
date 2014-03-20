package errors

import "fmt"

type HttpError interface {
	error
	StatusCode() int   // actual HTTP status code
	ErrorCode() string // error code returned in response body from CC or UAA
}

type httpError struct {
	statusCode  int
	code        string
	description string
}

type HttpNotFoundError struct {
	*httpError
}

func NewHttpError(statusCode int, code string, description string) HttpError {
	err := httpError{
		statusCode:  statusCode,
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
