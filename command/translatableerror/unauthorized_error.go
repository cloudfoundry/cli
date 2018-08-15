package translatableerror

const (
	BadCredentialMessage = "Bad credentials"
	InvalidOriginMessage = "The origin provided in the login hint is invalid."
)

type UnauthorizedError struct {
	Message string
}

func (e UnauthorizedError) Error() string {
	switch e.Message {
	case BadCredentialMessage:
		return "Credentials were rejected, please try again."
	case InvalidOriginMessage:
		return "The origin provided is invalid."
	default:
		return e.Message
	}
}

func (e UnauthorizedError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error())
}
