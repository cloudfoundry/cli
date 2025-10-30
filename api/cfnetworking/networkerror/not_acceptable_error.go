package networkerror

type NotAcceptableError struct {
	Message string `json:"error"`
}

func (e NotAcceptableError) Error() string {
	return e.Message
}
