package ccerror

// ServiceInstanceNameTakenError is returned when creating a
// service instance that already exists.
type ServiceInstanceNameTakenError struct {
	Message string
}

func (e ServiceInstanceNameTakenError) Error() string {
	return e.Message
}
