package actionerror

// ServiceInstanceParamsFetchingNotSupportedError is returned when
// service instance is user provided or
// service instance retrievable is false for a managed service instance
type ServiceInstanceParamsFetchingNotSupportedError struct {
}

func (e ServiceInstanceParamsFetchingNotSupportedError) Error() string {
	return "This service does not support fetching service instance parameters."
}
