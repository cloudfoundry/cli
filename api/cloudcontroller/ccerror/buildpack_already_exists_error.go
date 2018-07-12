package ccerror

type BuildpackAlreadyExistsError struct {
	Message string
}

func (e BuildpackAlreadyExistsError) Error() string {
	return e.Message
}
