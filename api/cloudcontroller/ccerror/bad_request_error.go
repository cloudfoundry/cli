package ccerror

// BadRequestError is returned when the server says the request was bad.
type BadRequestError struct {
	Message string
}

func (e BadRequestError) Error() string {
	return e.Message
}
