package translatableerror

type AccountLockedError struct {
	Message string
}

func (e AccountLockedError) Error() string {
	return e.Message
}

func (e AccountLockedError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error())
}
