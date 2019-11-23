package ccerror

type BuildpackInvalidError struct {
	Message string
}

func (e BuildpackInvalidError) Error() string {
	return e.Message
}
