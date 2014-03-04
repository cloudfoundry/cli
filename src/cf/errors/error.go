package errors

import (
	original "errors"
	"fmt"
)

type Error interface {
	error
	ErrorCode() string
}

type concreteError struct {
	Message   string
	errorCode string
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

func (err *concreteError) Error() string {
	return err.Message
}

func (err *concreteError) ErrorCode() string {
	return err.errorCode
}
