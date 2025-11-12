package networkerror

type UnauthorizedError struct {
	Message string `json:"error"`
}

func (e UnauthorizedError) Error() string {
	return e.Message
}
