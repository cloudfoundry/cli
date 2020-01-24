package ccerror

type QuotaAlreadyExists struct {
	Message string
}

func (e QuotaAlreadyExists) Error() string {
	return e.Message
}
