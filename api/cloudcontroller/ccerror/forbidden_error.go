package ccerror

// ForbiddenError is returned when the client is forbidden from executing the
// request.
type ForbiddenError struct {
	Message string
}

func (e ForbiddenError) Error() string {
	return e.Message
}
