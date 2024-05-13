package ccerror

// ServiceInstanceAlreadySharedError is returned when a
// service instance is already shared.
type ServiceInstanceAlreadySharedError struct {
	Message string
}

func (e ServiceInstanceAlreadySharedError) Error() string {
	return e.Message
}
