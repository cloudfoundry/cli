package networkerror

type ConflictError struct {
	Message string `json:"error"`
}

func (e ConflictError) Error() string {
	return e.Message
}
