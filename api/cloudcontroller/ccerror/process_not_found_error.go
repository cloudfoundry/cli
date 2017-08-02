package ccerror

// ProcessNotFoundError is returned when an endpoint cannot find the
// specified process
type ProcessNotFoundError struct {
}

func (e ProcessNotFoundError) Error() string {
	return "Process not found"
}
