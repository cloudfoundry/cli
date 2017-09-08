package actionerror

// StagingFailedError is returned when staging an application fails.
type StagingFailedError struct {
	Reason string
}

func (e StagingFailedError) Error() string {
	return e.Reason
}
