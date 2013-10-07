package net

import (
	"fmt"
)

type ApiStatus struct {
	Message    string
	ErrorCode  string
	StatusCode int

	isError    bool
	isNotFound bool
}

func NewApiStatus(message string, errorCode string, statusCode int) (apiStatus ApiStatus) {
	return ApiStatus{
		Message:    message,
		ErrorCode:  errorCode,
		StatusCode: statusCode,
		isError:    true,
	}
}

func NewApiStatusWithMessage(message string, a ...interface{}) (apiStatus ApiStatus) {
	return ApiStatus{
		Message: fmt.Sprintf(message, a...),
		isError: true,
	}
}

func NewApiStatusWithError(message string, err error) (apiStatus ApiStatus) {
	return ApiStatus{
		Message: fmt.Sprintf("%s: %s", message, err.Error()),
		isError: true,
	}
}

func NewNotFoundApiStatus(objectType string, objectId string) (apiStatus ApiStatus) {
	return ApiStatus{
		Message:    fmt.Sprintf("%s %s not found", objectType, objectId),
		isNotFound: true,
	}
}

func (apiStatus ApiStatus) IsError() bool {
	return apiStatus.isError
}

func (apiStatus ApiStatus) IsNotFound() bool {
	return apiStatus.isNotFound
}

func (apiStatus ApiStatus) Successful() bool {
	return !apiStatus.NotSuccessful()
}

func (apiStatus ApiStatus) NotSuccessful() bool {
	return apiStatus.IsError() || apiStatus.IsNotFound()
}
