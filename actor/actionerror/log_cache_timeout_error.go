package actionerror

type LogCacheTimeoutError struct{}

func (LogCacheTimeoutError) Error() string {
	return "Timeout trying to connect to LogCache"
}
