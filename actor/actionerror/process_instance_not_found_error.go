package actionerror

// ProcessInstanceNotFoundError is returned when trying
// to ssh into an instance that doesn't exist
type ProcessInstanceNotFoundError struct {
}

func (e ProcessInstanceNotFoundError) Error() string {
	return "The specified application instance does not exist"
}
