package ccerror

import "fmt"

// APINotFoundError is returned when the API endpoint is not found.
type APINotFoundError struct {
	URL string
}

func (e APINotFoundError) Error() string {
	return fmt.Sprintf("Unable to find API at %s", e.URL)
}
