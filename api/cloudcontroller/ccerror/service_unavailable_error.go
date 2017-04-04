package ccerror

// ServiceUnavailableError wraps a http 503 error.
type ServiceUnavailableError struct {
	Message string
}

func (e ServiceUnavailableError) Error() string {
	return e.Message
}
