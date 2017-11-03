package translatableerror

type RoutePathWithTCPDomainError struct{}

func (RoutePathWithTCPDomainError) Error() string {
	return "The route is invalid: a route path cannot be used with a TCP domain."
}

func (e RoutePathWithTCPDomainError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error())
}
