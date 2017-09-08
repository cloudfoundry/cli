package actionerror

// HTTPHealthCheckInvalidError is returned when an HTTP endpoint is used with a
// health check type that is not HTTP.
type HTTPHealthCheckInvalidError struct {
}

func (e HTTPHealthCheckInvalidError) Error() string {
	return "Health check type must be 'http' to set a health check HTTP endpoint"
}
