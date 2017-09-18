package networkerror

type ForbiddenError struct {
	Message string `json:"error"`
}

func (e ForbiddenError) Error() string {
	return e.Message
}
