package ccerror

// ApplicationStoppedStatsError is returned when requesting instance
// information from a stopped app.
type ApplicationStoppedStatsError struct {
	Message string
}

func (e ApplicationStoppedStatsError) Error() string {
	return e.Message
}
