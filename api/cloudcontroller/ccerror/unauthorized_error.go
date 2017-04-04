package ccerror

// UnauthorizedError is returned when the client does not have the correct
// permissions to execute the request.
type UnauthorizedError struct {
	Message string
}

func (e UnauthorizedError) Error() string {
	return e.Message
}
