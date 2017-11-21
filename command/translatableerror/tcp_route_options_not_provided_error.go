package translatableerror

type TCPRouteOptionsNotProvidedError struct {
}

func (e TCPRouteOptionsNotProvidedError) Error() string {
	return "The route is invalid: For TCP routes you must specify a port or request a random one."
}

func (e TCPRouteOptionsNotProvidedError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), nil)
}
