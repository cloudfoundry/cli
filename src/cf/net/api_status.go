package net

import "fmt"

const (
	ORG_EXISTS                  = "30002"
	SPACE_EXISTS                = "40002"
	APP_NOT_STAGED              = "170002"
	SERVICE_INSTANCE_NAME_TAKEN = "60002"
)

type ApiStatus struct {
	Message    string
	ErrorCode  string
	StatusCode int

	isError bool
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

func (apiStatus ApiStatus) IsError() bool {
	return apiStatus.isError
}
