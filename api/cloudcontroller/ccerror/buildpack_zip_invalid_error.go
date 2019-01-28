package ccerror

type BuildpackZipInvalidError struct {
	Message string
}

func (e BuildpackZipInvalidError) Error() string {
	return e.Message
}
