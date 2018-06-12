package ccerror

import (
	"fmt"
	"strings"
)

type MultiError struct {
	Errors       []V3Error
	ResponseCode int
}

func (e MultiError) Details() []string {
	var errorMsg []string

	for _, err := range e.Errors {
		errorMsg = append(errorMsg, err.Detail)
	}

	return errorMsg
}

func (e MultiError) Error() string {
	errorMsg := []string{
		"Multiple errors occurred:",
		fmt.Sprintf("Response Code: %d", e.ResponseCode),
	}

	for _, err := range e.Errors {
		errorMsg = append(errorMsg, fmt.Sprintf("Code: %d, Title: %s, Detail: %s", err.Code, err.Title, err.Detail))
	}

	return strings.Join(errorMsg, "\n")
}
