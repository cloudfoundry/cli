package actionerror

// ServiceInstanceAlreadySharedError is returned when attempting to shared a
// service instance to a space in which the service instance has already been
// shared with.
type ServiceInstanceAlreadySharedError struct{}

func (e ServiceInstanceAlreadySharedError) Error() string {
	return "The service instance has already been shared with this space"
}
