package ccerror

// ServiceInstanceParametersFetchNotSupportedError is returned when
// the service instance is user-provided or
// service instance fetching is not supported for managed instances
type ServiceInstanceParametersFetchNotSupportedError struct {
	Message string
}

func (e ServiceInstanceParametersFetchNotSupportedError) Error() string {
	return e.Message
}
