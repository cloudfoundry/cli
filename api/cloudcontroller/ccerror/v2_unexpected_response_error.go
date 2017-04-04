package ccerror

import "fmt"

// V2ErrorResponse represents a generic Cloud Controller V2 error response.
type V2ErrorResponse struct {
	Code        int    `json:"code"`
	Description string `json:"description"`
	ErrorCode   string `json:"error_code"`
}

// V2UnexpectedResponseError is returned when the client gets an error that has
// not been accounted for.
type V2UnexpectedResponseError struct {
	V2ErrorResponse

	RequestIDs   []string
	ResponseCode int
}

func (e V2UnexpectedResponseError) Error() string {
	message := fmt.Sprintf("Unexpected Response\nResponse code: %d\nCC code:       %d\nCC error code: %s", e.ResponseCode, e.Code, e.ErrorCode)
	for _, id := range e.RequestIDs {
		message = fmt.Sprintf("%s\nRequest ID:    %s", message, id)
	}
	return fmt.Sprintf("%s\nDescription:   %s", message, e.Description)
}
