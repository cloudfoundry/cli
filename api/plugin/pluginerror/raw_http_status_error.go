package pluginerror

import "fmt"

// RawHTTPStatusError represents any response with a 4xx or 5xx status code.
type RawHTTPStatusError struct {
	Status      string
	RawResponse []byte
}

func (r RawHTTPStatusError) Error() string {
	return fmt.Sprintf("HTTP Response: %s\nHTTP Response Body: %s", r.Status, r.RawResponse)
}
