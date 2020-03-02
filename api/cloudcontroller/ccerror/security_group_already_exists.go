package ccerror

type SecurityGroupAlreadyExists struct {
	Message string
}

func (e SecurityGroupAlreadyExists) Error() string {
	return e.Message
}
