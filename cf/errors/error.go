package errors

import (
	original "errors"
	"fmt"
)

func New(message string) error {
	return original.New(message)
}

func NewWithError(message string, err error) error {
	return fmt.Errorf("%s: %s", message, err.Error())
}

func NewWithSlice(errs []error) error {
	message := ""

	for _, err := range errs {
		message = fmt.Sprintf("%s%s\n", message, err.Error())
	}
	return New(message)
}
