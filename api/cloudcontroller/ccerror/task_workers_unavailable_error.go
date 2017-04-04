package ccerror

// TaskWorkersUnavailableError represents the case when no Diego workers are
// available.
type TaskWorkersUnavailableError struct {
	Message string
}

func (e TaskWorkersUnavailableError) Error() string {
	return e.Message
}
