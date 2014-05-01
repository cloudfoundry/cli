package errors

import (
	original "errors"
	"fmt"
)

func New(message string) error {
	return original.New(message)
}

func NewWithFmt(message string, args ...interface{}) error {
	return original.New(fmt.Sprintf(message, args...))
}

func NewWithError(message string, err error) error {
	return NewWithFmt("%s: %s", message, err.Error())
}

func NewWithSlice(errs []error) error {
	message := ""

	for _, err := range errs {
		message = fmt.Sprintf("%s%s\n", message, err.Error())
	}
	return New(message)
}
