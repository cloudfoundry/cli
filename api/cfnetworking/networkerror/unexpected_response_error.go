package networkerror

import "fmt"

// UnexpectedResponseError is returned when the client gets an error that has
// not been accounted for.
type UnexpectedResponseError struct {
	ErrorResponse

	RequestIDs   []string
	ResponseCode int
}

func (e UnexpectedResponseError) Error() string {
	message := fmt.Sprintf("Unexpected Response\nResponse code: %d", e.ResponseCode)
	for _, id := range e.RequestIDs {
		message = fmt.Sprintf("%s\nRequest ID:    %s", message, id)
	}
	return fmt.Sprintf("%s\nDescription:   %s", message, e.Message)
}
