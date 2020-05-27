package translatableerror

type PasswordVerificationFailedError struct{}

func (PasswordVerificationFailedError) Error() string {
	return "Password verification does not match."
}

func (e PasswordVerificationFailedError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error())
}
