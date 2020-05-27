package ccerror

type SecurityGroupNotBound struct {
	Message string
}

func (e SecurityGroupNotBound) Error() string {
	return e.Message
}
