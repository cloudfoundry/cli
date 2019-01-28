package ccerror

type BuildpackStacksDontMatchError struct {
	Message string
}

func (e BuildpackStacksDontMatchError) Error() string {
	return e.Message
}
