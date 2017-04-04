package ccerror

// InstancesError is returned when requesting instance information encounters
// an error.
type InstancesError struct {
	Message string
}

func (e InstancesError) Error() string {
	return e.Message
}
