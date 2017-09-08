package actionerror

// StagingFailedNoAppDetectedError is returned when staging an application fails.
type StagingFailedNoAppDetectedError struct {
	Reason string
}

func (e StagingFailedNoAppDetectedError) Error() string {
	return e.Reason
}
