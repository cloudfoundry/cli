package ccerror

// UnprocessableEntityError is returned when the request cannot be processed by
// the cloud controller.
type UnprocessableEntityError struct {
	Message string
}

func (e UnprocessableEntityError) Error() string {
	return e.Message
}
