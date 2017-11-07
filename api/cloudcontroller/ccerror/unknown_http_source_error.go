package ccerror

import "fmt"

// UnknownHTTPSourceError represents HTTP responses with status code >= 400
// that we cannot unmarshal into a {V2,V3}ErrorResponse.
// Ex: In "cf api google.com", google will return a 404, but the body will not match
// an error from the Cloud Controller
type UnknownHTTPSourceError struct {
	StatusCode  int
	RawResponse []byte
}

func (r UnknownHTTPSourceError) Error() string {
	return fmt.Sprintf("Error unmarshalling the following into a cloud controller error: %s", r.RawResponse)
}
