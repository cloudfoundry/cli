package translatableerror

type MissingCredentialsError struct {
	MissingUsername bool
	MissingPassword bool
}

func (MissingCredentialsError) DisplayUsage() {}

func (e MissingCredentialsError) Error() string {
	switch {
	case e.MissingUsername && !e.MissingPassword:
		return "Username not provided."
	case !e.MissingUsername && e.MissingPassword:
		return "Password not provided."
	default:
		return "Username and password not provided."
	}
}

func (e MissingCredentialsError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error())
}
