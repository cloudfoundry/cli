package actionerror

// SharedServiceInstanceNotFoundError is returned when a service instance is not found when performing a share service.
type SharedServiceInstanceNotFoundError struct {
}

func (e SharedServiceInstanceNotFoundError) Error() string {
	return "Specified instance not found or not a managed service instance. Sharing is not supported for user provided services."
}
