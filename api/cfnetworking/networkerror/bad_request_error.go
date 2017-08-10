package networkerror

type BadRequestError struct {
	Message string `json:"error"`
}

func (e BadRequestError) Error() string {
	return e.Message
}
