package ccerror

// InstanceNotFoundError is returned when an endpoint cannot find the
// specified instance
type InstanceNotFoundError struct {
}

func (e InstanceNotFoundError) Error() string {
	return "Instance not found"
}
