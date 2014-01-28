package net

import (
	"fmt"
)

type ApiResponse struct {
	Message     string
	ErrorCode   string
	StatusCode  int
	ErrorHeader string
	ErrorBody   string

	isError        bool
	isHttpResponse bool
	isNotFound     bool
}

func NewApiResponse(message string, errorCode string, statusCode int) (apiResponse ApiResponse) {
	return ApiResponse{
		Message:        message,
		ErrorCode:      errorCode,
		StatusCode:     statusCode,
		isError:        true,
		isHttpResponse: true,
	}
}

func NewApiResponseWithHttpError(message string, errorCode string, statusCode int, errorHeader string, errorBody string) (apiResponse ApiResponse) {
	return ApiResponse{
		Message:        message,
		ErrorCode:      errorCode,
		StatusCode:     statusCode,
		isError:        true,
		isHttpResponse: true,
		ErrorHeader:    errorHeader,
		ErrorBody:      errorBody,
	}
}

func NewApiResponseWithStatusCode(statusCode int) (apiResponse ApiResponse) {
	return ApiResponse{
		StatusCode:     statusCode,
		isHttpResponse: true,
	}
}

func NewApiResponseWithMessage(message string, a ...interface{}) (apiResponse ApiResponse) {
	return ApiResponse{
		Message: fmt.Sprintf(message, a...),
		isError: true,
	}
}

func NewApiResponseWithError(message string, err error) (apiResponse ApiResponse) {
	return ApiResponse{
		Message: fmt.Sprintf("%s: %s", message, err.Error()),
		isError: true,
	}
}

func NewNotFoundApiResponse(message string, a ...interface{}) (apiResponse ApiResponse) {
	return ApiResponse{
		Message:    fmt.Sprintf(message, a...),
		isNotFound: true,
	}
}

func NewSuccessfulApiResponse() (apiResponse ApiResponse) {
	return ApiResponse{}
}

func (apiResponse ApiResponse) IsError() bool {
	return apiResponse.isError
}

func (apiResponse ApiResponse) IsHttpError() bool {
	return apiResponse.isError && apiResponse.isHttpResponse
}

func (apiResponse ApiResponse) IsNotFound() bool {
	return apiResponse.isNotFound || (apiResponse.isHttpResponse && apiResponse.StatusCode == 404)
}

func (apiResponse ApiResponse) IsSuccessful() bool {
	return !apiResponse.IsNotSuccessful()
}

func (apiResponse ApiResponse) IsNotSuccessful() bool {
	return apiResponse.IsError() || apiResponse.IsNotFound()
}
