package actionerror

type TCPRouteOptionsNotProvidedError struct {
	Domain string
}

func (e TCPRouteOptionsNotProvidedError) Error() string {
	return "The route is invalid: For TCP routes you must specify a port or request a random one."
}
