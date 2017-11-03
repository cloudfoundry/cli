package actionerror

type RoutePathWithTCPDomainError struct {
}

func (RoutePathWithTCPDomainError) Error() string {
	return "cannot use provided route path with a TCP domain"
}
