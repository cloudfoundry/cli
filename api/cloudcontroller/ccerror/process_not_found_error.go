package ccerror

// ProcessNotFoundError is returned when the API endpoint is not found.
type ProcessNotFoundError struct {
}

func (e ProcessNotFoundError) Error() string {
	return "Process not found"
}
