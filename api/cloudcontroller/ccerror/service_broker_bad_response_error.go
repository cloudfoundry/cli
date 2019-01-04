package ccerror

// ServiceBrokerBadResponseError is returned when CC returns an error
// of type CF-ServiceBrokerBadResponse.
type ServiceBrokerBadResponseError struct {
	Message string
}

func (e ServiceBrokerBadResponseError) Error() string {
	return e.Message
}
