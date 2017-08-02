package ccerror

// ApplicationNotFoundError is returned when an endpoint cannot find the
// specified application
type ApplicationNotFoundError struct {
}

func (e ApplicationNotFoundError) Error() string {
	return "Application not found"
}
