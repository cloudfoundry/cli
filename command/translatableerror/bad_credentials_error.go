package translatableerror

type BadCredentialsError struct{}

func (BadCredentialsError) Error() string {
	return "Credentials were rejected, please try again."
}

func (e BadCredentialsError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{})
}
