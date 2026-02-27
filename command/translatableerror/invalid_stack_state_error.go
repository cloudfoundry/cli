package translatableerror

type InvalidStackStateError struct {
	State string
}

func (InvalidStackStateError) Error() string {
	return "Invalid stack state: {{.State}}. Expected one of: ACTIVE, RESTRICTED, DEPRECATED, DISABLED"
}

func (e InvalidStackStateError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"State": e.State,
	})
}
