package actionerror

// TaskWorkersUnavailableError is returned when there are no workers to run a
// given task.
type TaskWorkersUnavailableError struct {
	Message string
}

func (e TaskWorkersUnavailableError) Error() string {
	return e.Message
}
