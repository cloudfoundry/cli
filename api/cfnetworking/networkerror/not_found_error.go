package networkerror

// NotFoundError wraps a generic 404 error.
type NotFoundError struct {
	Message string
}

func (e NotFoundError) Error() string {
	return e.Message
}
