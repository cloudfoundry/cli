package translatableerror

type SSHUnableToAuthenticateError struct{}

func (SSHUnableToAuthenticateError) Error() string {
	return "Error opening SSH connection: You are not authorized to perform the requested action."
}

func (e SSHUnableToAuthenticateError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error())
}
