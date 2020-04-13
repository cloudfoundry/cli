package app

import (
	"fmt"
	"net/http"
)

type notFoundError struct{}

func (notFoundError) Error() string {
	return http.StatusText(http.StatusNotFound)
}

func (notFoundError) StatusCode() int {
	return http.StatusNotFound
}

type unauthorizedError struct{}

func (unauthorizedError) Error() string {
	return http.StatusText(http.StatusUnauthorized)
}

func (unauthorizedError) StatusCode() int {
	return http.StatusUnauthorized
}

type badRequestError struct {
	message string
	err     error
}

func newBadRequestError(message string, err error) error {
	return &badRequestError{
		message: message,
		err:     err,
	}
}

func (e *badRequestError) Error() string {
	return fmt.Sprintf("%s: %s", e.message, e.err.Error())
}

func (badRequestError) StatusCode() int {
	return http.StatusBadRequest
}
