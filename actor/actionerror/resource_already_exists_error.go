package actionerror

// ResourceAlreadyExistsError is returned when an identical resource already exists
type ResourceAlreadyExistsError struct {
	Message string
}

func (e ResourceAlreadyExistsError) Error() string {
	return e.Message
}
