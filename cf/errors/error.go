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
