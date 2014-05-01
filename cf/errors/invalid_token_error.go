package errors

type InvalidTokenError struct {
	description string
}

func NewInvalidTokenError(description string) error {
	return &InvalidTokenError{description: description}
}

func (err *InvalidTokenError) Error() string {
	return "Invalid auth token: " + err.description
}
