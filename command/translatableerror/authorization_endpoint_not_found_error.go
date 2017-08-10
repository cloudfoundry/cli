package translatableerror

type AuthorizationEndpointNotFoundError struct {
}

func (AuthorizationEndpointNotFoundError) Error() string {
	return "No Authorization Endpoint Found"
}

func (e AuthorizationEndpointNotFoundError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error())
}
