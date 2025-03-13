package networkerror

type ErrorResponse struct {
	Message string `json:"error"`
}

func (e ErrorResponse) Error() string {
	return e.Message
}
