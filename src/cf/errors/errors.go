package errors

import (
	original "errors"
	"fmt"
)

type Error interface {
	error
	IsHttpError() bool
	IsNotFound() bool
	StatusCode() int
	ErrorCode() string
}

type HttpError interface {
	Error
	Headers() string
	Body() string
}

type concreteError struct {
	Message   string
	errorCode string

	statusCode  int
	errorHeader string
	errorBody   string

	isHttpResponse bool
	isNotFound     bool
}

func New(message string) error {
	return original.New(message)
}

func NewError(message string, errorCode string, statusCode int) Error {
	return &concreteError{
		Message:        message,
		errorCode:      errorCode,
		statusCode:     statusCode,
		isHttpResponse: true,
	}
}

func NewErrorWithHttpError(message string, errorCode string, statusCode int, errorHeader string, errorBody string) Error {
	return &concreteError{
		Message:        message,
		errorCode:      errorCode,
		statusCode:     statusCode,
		isHttpResponse: true,
		errorHeader:    errorHeader,
		errorBody:      errorBody,
	}
}

func NewHttpErrorWithHttpError(message string, errorCode string, statusCode int, errorHeader string, errorBody string) HttpError {
	return &concreteError{
		Message:        message,
		errorCode:      errorCode,
		statusCode:     statusCode,
		isHttpResponse: true,
		errorHeader:    errorHeader,
		errorBody:      errorBody,
	}
}

func NewErrorWithStatusCode(statusCode int) Error {
	return &concreteError{
		statusCode:     statusCode,
		isHttpResponse: true,
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

func NewHTTPErrorWithError(message string, err error) HttpError {
	return &concreteError{
		Message: fmt.Sprintf("%s: %s", message, err.Error()),
		isHttpResponse: true,
		statusCode:
	}
}

func NewNotFoundError(message string, a ...interface{}) Error {
	return &concreteError{
		Message:    fmt.Sprintf(message, a...),
		isNotFound: true,
	}
}

func NewSuccessfulError() (error Error) {
	return &concreteError{}
}

func (err *concreteError) IsHttpError() bool {
	return err.isHttpResponse
}

func (err *concreteError) IsNotFound() bool {
	return err.isNotFound || (err.isHttpResponse && err.statusCode == 404)
}

func (err *concreteError) Error() string {
	return err.Message
}

func (err *concreteError) StatusCode() int {
	return err.statusCode
}
func (err *concreteError) ErrorCode() string {
	return err.errorCode
}

func (err *concreteError) Headers() string {
	return err.errorHeader
}

func (err *concreteError) Body() string {
	return err.errorBody
}
