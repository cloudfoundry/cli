package actionerror

import "fmt"

// RouteNotFoundError is returned when a route cannot be found
type RouteNotFoundError struct {
	Host       string
	DomainGUID string
	Path       string
	Port       int
}

func (e RouteNotFoundError) Error() string {
	if e.Path != "" {
		return fmt.Sprintf("Route with host %s, domain guid %s, and path %s not found", e.Host, e.DomainGUID, e.Path)
	}
	return fmt.Sprintf("Route with host %s and domain guid %s not found", e.Host, e.DomainGUID)
}
