package ccerror

import (
	"fmt"
	"strings"
)

// V3ErrorResponse represents a generic Cloud Controller V3 error response.
type V3ErrorResponse struct {
	Errors []V3Error `json:"errors"`
}

// V3Error represents a cloud controller error.
type V3Error struct {
	Code   int    `json:"code"`
	Detail string `json:"detail"`
	Title  string `json:"title"`
}

// V3UnexpectedResponseError is returned when the client gets an error that has
// not been accounted for.
type V3UnexpectedResponseError struct {
	V3ErrorResponse

	ResponseCode int
	RequestIDs   []string
}

func (e V3UnexpectedResponseError) Error() string {
	messages := []string{
		"Unexpected Response",
		fmt.Sprintf("Response Code: %d", e.ResponseCode),
	}

	for _, id := range e.RequestIDs {
		messages = append(messages, fmt.Sprintf("Request ID:    %s", id))
	}

	for _, ccError := range e.V3ErrorResponse.Errors {
		messages = append(messages, fmt.Sprintf("Code: %d, Title: %s, Detail: %s", ccError.Code, ccError.Title, ccError.Detail))
	}

	return strings.Join(messages, "\n")
}
