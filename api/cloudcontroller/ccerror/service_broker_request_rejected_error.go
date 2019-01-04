package ccerror

// ServiceBrokerRequestRejectedError is returned when CC returns an error
// of type CF-ServiceBrokerRequestRejected.
type ServiceBrokerRequestRejectedError struct {
	Message string
}

func (e ServiceBrokerRequestRejectedError) Error() string {
	return e.Message
}
