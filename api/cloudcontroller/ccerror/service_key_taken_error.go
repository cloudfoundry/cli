package ccerror

// ServiceKeyTakenError is returned when creating a
// service key that already exists
type ServiceKeyTakenError struct {
	Message string
}

func (e ServiceKeyTakenError) Error() string {
	return e.Message
}
