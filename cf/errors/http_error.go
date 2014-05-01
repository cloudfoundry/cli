package errors

import "fmt"

type HttpError interface {
	error
	StatusCode() int   // actual HTTP status code
	ErrorCode() string // error code returned in response body from CC or UAA
}

type baseHttpError struct {
	statusCode   int
	apiErrorCode string
	description  string
}

type HttpNotFoundError struct {
	baseHttpError
}

func NewHttpError(statusCode int, code string, description string) error {
	err := baseHttpError{
		statusCode:   statusCode,
		apiErrorCode: code,
		description:  description,
	}
	switch statusCode {
	case 404:
		return &HttpNotFoundError{err}
	default:
		return &err
	}
}

func (err *baseHttpError) StatusCode() int {
	return err.statusCode
}

func (err *baseHttpError) Error() string {
	return fmt.Sprintf(
		"Server error, status code: %d, error code: %s, message: %s",
		err.statusCode,
		err.apiErrorCode,
		err.description,
	)
}

func (err *baseHttpError) ErrorCode() string {
	return err.apiErrorCode
}
