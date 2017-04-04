package ccerror

// ResourceNotFoundError is returned when the client requests a resource that
// does not exist or does not have permissions to see.
type ResourceNotFoundError struct {
	Message string
}

func (e ResourceNotFoundError) Error() string {
	return e.Message
}
