package ccerror

// ServiceInstanceOperationInProgressError is returned when an operation
// cannot proceed because an operation is already in progress
type ServiceInstanceOperationInProgressError struct {
	Message string
}

func (e ServiceInstanceOperationInProgressError) Error() string {
	return e.Message
}
