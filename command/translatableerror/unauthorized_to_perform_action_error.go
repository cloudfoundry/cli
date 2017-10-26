package translatableerror

type UnauthorizedToPerformActionError struct {
}

func (UnauthorizedToPerformActionError) Error() string {
	return "You are not authorized to perform the requested action."
}

func (e UnauthorizedToPerformActionError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), nil)
}
