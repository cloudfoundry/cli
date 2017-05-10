package pluginerror

import "fmt"

// RawHTTPStatusError represents any response with a 4xx or 5xx status code.
type RawHTTPStatusError struct {
	Status      string
	RawResponse []byte
}

func (r RawHTTPStatusError) Error() string {
	return fmt.Sprintf("Error Code: %s\nRaw Response: %s", r.Status, r.RawResponse)
}
