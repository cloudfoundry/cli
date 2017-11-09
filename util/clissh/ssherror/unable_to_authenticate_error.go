package ssherror

type UnableToAuthenticateError struct {
	Err error
}

func (e UnableToAuthenticateError) Error() string {
	return e.Err.Error()
}
