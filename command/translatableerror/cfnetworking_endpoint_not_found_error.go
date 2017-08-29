package translatableerror

type CFNetworkingEndpointNotFoundError struct {
}

func (CFNetworkingEndpointNotFoundError) Error() string {
	return "This command requires Network Policy API V1. Your targeted endpoint does not expose it."
}

func (e CFNetworkingEndpointNotFoundError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error())
}
