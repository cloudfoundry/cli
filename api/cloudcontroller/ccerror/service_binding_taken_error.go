package ccerror

// ServiceBindingTakenError is returned when creating a
// service binding that already exists
type ServiceBindingTakenError struct {
	Message string
}

func (e ServiceBindingTakenError) Error() string {
	return e.Message
}
