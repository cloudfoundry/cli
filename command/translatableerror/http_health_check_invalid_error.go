package translatableerror

type HTTPHealthCheckInvalidError struct {
}

func (HTTPHealthCheckInvalidError) Error() string {
	return "Health check type must be 'http' to set a health check HTTP endpoint."
}

func (e HTTPHealthCheckInvalidError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error())
}
