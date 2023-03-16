package actionerror

import "fmt"

// RouteNotFoundError is returned when a route cannot be found
type RouteNotFoundError struct {
	Host       string
	DomainName string
	Path       string
	Port       int
}

func (e RouteNotFoundError) Error() string {
	switch e.Port {
	case 0:
		return fmt.Sprintf("Route with host '%s', domain '%s', and path '%s' not found.", e.Host, e.DomainName, e.path())
	default:
		return fmt.Sprintf("Route with domain '%s' and port %d not found.", e.DomainName, e.Port)
	}
}

func (e RouteNotFoundError) path() string {
	if e.Path == "" {
		return "/"
	}
	return e.Path
}
