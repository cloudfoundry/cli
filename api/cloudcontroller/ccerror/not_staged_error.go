package ccerror

// NotStagedError is returned when requesting instance information from a
// not staged app.
type NotStagedError struct {
	Message string
}

func (e NotStagedError) Error() string {
	return e.Message
}
