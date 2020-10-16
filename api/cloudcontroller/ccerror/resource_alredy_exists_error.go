package ccerror

// ResourceAlreadyExistsError is returned when a resource cannot be created
// because an identical resource already exists in the Cloud Controller.
type ResourceAlreadyExistsError struct {
	Message string
}

func (e ResourceAlreadyExistsError) Error() string {
	return e.Message
}
