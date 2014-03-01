package errors

import (
	original "errors"
	"fmt"
)

type Error interface {
	error
	IsNotFound() bool
	ErrorCode() string
}

type concreteError struct {
	Message    string
	errorCode  string
	isNotFound bool
}

func New(message string) error {
	return original.New(message)
}

func NewError(message string, errorCode string) Error {
	return &concreteError{
		Message:   message,
		errorCode: errorCode,
	}
}

func NewErrorWithMessage(message string, a ...interface{}) Error {
	return &concreteError{
		Message: fmt.Sprintf(message, a...),
	}
}

func NewErrorWithError(message string, err error) Error {
	return &concreteError{
		Message: fmt.Sprintf("%s: %s", message, err.Error()),
	}
}

func NewNotFoundError(message string, a ...interface{}) Error {
	return &concreteError{
		Message:    fmt.Sprintf(message, a...),
		isNotFound: true,
	}
}

func (err *concreteError) IsNotFound() bool {
	return err.isNotFound
}

func (err *concreteError) Error() string {
	return err.Message
}

func (err *concreteError) ErrorCode() string {
	return err.errorCode
}

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

func NewHttpError(statusCode int, header string, body string, code string, description string) HttpError {
	return &httpError{
		statusCode:  statusCode,
		headers:     header,
		body:        body,
		code:        code,
		description: description,
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

func (err *httpError) IsNotFound() bool {
	return err.statusCode == 404
}
