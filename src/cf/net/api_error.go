package net

import "fmt"

const(
	ORG_EXISTS = "30002"
)

type ApiError struct {
	Message    string
	ErrorCode  string
	StatusCode int
}

func NewApiError(message string, errorCode string, statusCode int) (apiErr *ApiError) {
	return &ApiError{
		Message:    message,
		ErrorCode:  errorCode,
		StatusCode: statusCode,
	}
}

func NewApiErrorWithMessage(message string, a ...interface{}) (apiErr *ApiError) {
	return &ApiError{
		Message: fmt.Sprintf(message, a...),
	}
}

func NewApiErrorWithError(message string, err error) (apiErr *ApiError) {
	return &ApiError{
		Message: fmt.Sprintf("%s: %s", message, err.Error()),
	}
}

func (apiErr *ApiError) Error() string {
	return apiErr.Message
}
