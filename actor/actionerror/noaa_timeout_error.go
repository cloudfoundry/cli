package actionerror

type NOAATimeoutError struct{}

func (NOAATimeoutError) Error() string {
	return "Timeout trying to connect to NOAA"
}
