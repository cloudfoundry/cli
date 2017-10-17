package actionerror

// SSHEndpointNotSetError is returned when staging an application fails.
type SSHEndpointNotSetError struct {
}

func (e SSHEndpointNotSetError) Error() string {
	return "SSH endpoint not set"
}
