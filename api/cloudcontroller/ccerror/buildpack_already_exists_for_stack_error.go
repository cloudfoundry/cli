package ccerror

type BuildpackAlreadyExistsForStackError struct {
	Message string
}

func (e BuildpackAlreadyExistsForStackError) Error() string {
	return e.Message
}
