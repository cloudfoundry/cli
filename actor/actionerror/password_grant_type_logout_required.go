package actionerror

// PasswordGrantTypeLogoutRequiredError is returned when a user tries to
// authenticate with the password grant type, but a previous user had
// authenticated with the client grant type
type PasswordGrantTypeLogoutRequiredError struct{}

func (PasswordGrantTypeLogoutRequiredError) Error() string {
	return "Service account currently logged in. Use 'cf logout' to log out service account and try again."
}
