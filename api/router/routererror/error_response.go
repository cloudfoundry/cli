package routererror

import "fmt"

type ErrorResponse struct {
	Name       string `json:"name"`
	Message    string `json:"message"`
	StatusCode int
}

func (err ErrorResponse) Error() string {
	return fmt.Sprintf("Server error, status code: %d, error code: %s, message: %s", err.StatusCode, err.Name, err.Message)
}
