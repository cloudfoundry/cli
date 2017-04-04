package ccerror

import "fmt"

// RawHTTPStatusError represents any response with a 4xx or 5xx status code.
type RawHTTPStatusError struct {
	StatusCode  int
	RawResponse []byte
	RequestIDs  []string
}

func (r RawHTTPStatusError) Error() string {
	return fmt.Sprintf("Error Code: %d\nRaw Response: %s", r.StatusCode, r.RawResponse)
}
