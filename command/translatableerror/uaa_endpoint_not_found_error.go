package translatableerror

type UAAEndpointNotFoundError struct {
}

func (UAAEndpointNotFoundError) Error() string {
	return "No UAA Endpoint Found"
}

func (e UAAEndpointNotFoundError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error())
}
