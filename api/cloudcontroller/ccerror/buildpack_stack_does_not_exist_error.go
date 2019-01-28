package ccerror

type BuildpackStackDoesNotExistError struct {
	Message string
}

func (e BuildpackStackDoesNotExistError) Error() string {
	return e.Message
}
