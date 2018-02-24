package translatableerror

type PasswordGrantTypeLogoutRequiredError struct{}

func (PasswordGrantTypeLogoutRequiredError) Error() string {
	return "Service account currently logged in. Use 'cf logout' to log out service account and try again."
}

func (e PasswordGrantTypeLogoutRequiredError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error())
}
