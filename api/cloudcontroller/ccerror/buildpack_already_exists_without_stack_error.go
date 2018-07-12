package ccerror

type BuildpackAlreadyExistsWithoutStackError struct {
	Message string
}

func (e BuildpackAlreadyExistsWithoutStackError) Error() string {
	return e.Message
}
