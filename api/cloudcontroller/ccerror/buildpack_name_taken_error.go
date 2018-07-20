package ccerror

type BuildpackNameTakenError struct {
	Message string
}

func (e BuildpackNameTakenError) Error() string {
	return e.Message
}
