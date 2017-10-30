package actionerror

// NotLoggedInError represents the scenario when the user is not logged in.
type NotLoggedInError struct {
	BinaryName string
}

func (NotLoggedInError) Error() string {
	// The error message will be replaced by a translated message, returning the
	// empty string does not add to the translation files.
	return ""
}
