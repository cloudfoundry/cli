package net

import (
	"fmt"
)

type ApiResponse struct {
	Message    string
	ErrorCode  string
	StatusCode int

	isError    bool
	isNotFound bool
}

func NewApiStatus(message string, errorCode string, statusCode int) (apiResponse ApiResponse) {
	return ApiResponse{
		Message:    message,
		ErrorCode:  errorCode,
		StatusCode: statusCode,
		isError:    true,
	}
}

func NewApiStatusWithMessage(message string, a ...interface{}) (apiResponse ApiResponse) {
	return ApiResponse{
		Message: fmt.Sprintf(message, a...),
		isError: true,
	}
}

func NewApiStatusWithError(message string, err error) (apiResponse ApiResponse) {
	return ApiResponse{
		Message: fmt.Sprintf("%s: %s", message, err.Error()),
		isError: true,
	}
}

func NewNotFoundApiStatus(objectType string, objectId string) (apiResponse ApiResponse) {
	return ApiResponse{
		Message:    fmt.Sprintf("%s %s not found", objectType, objectId),
		isNotFound: true,
	}
}

func (apiResponse ApiResponse) IsError() bool {
	return apiResponse.isError
}

func (apiResponse ApiResponse) IsNotFound() bool {
	return apiResponse.isNotFound
}

func (apiResponse ApiResponse) IsSuccessful() bool {
	return !apiResponse.IsNotSuccessful()
}

func (apiResponse ApiResponse) IsNotSuccessful() bool {
	return apiResponse.IsError() || apiResponse.IsNotFound()
}
