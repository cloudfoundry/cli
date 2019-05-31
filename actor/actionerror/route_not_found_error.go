package actionerror

import "fmt"

// RouteNotFoundError is returned when a route cannot be found
type RouteNotFoundError struct {
	Host       string
	DomainGUID string
	DomainName string
	Path       string
	Port       int
}

func (e RouteNotFoundError) Error() string {
	if e.DomainName != "" {
		switch {
		case e.Host != "" && e.Path != "":
			return fmt.Sprintf("Route with host '%s', domain '%s', and path '%s' not found.", e.Host, e.DomainName, e.Path)
		case e.Host != "":
			return fmt.Sprintf("Route with host '%s' and domain '%s' not found.", e.Host, e.DomainName)
		case e.Path != "":
			return fmt.Sprintf("Route with domain '%s' and path '%s' not found.", e.DomainName, e.Path)
		default:
			return fmt.Sprintf("Route with domain '%s' not found.", e.DomainName)
		}
	}
	if e.Path != "" {
		return fmt.Sprintf("Route with host '%s', domain guid '%s', and path '%s' not found.", e.Host, e.DomainGUID, e.Path)
	}
	return fmt.Sprintf("Route with host '%s' and domain guid '%s' not found.", e.Host, e.DomainGUID)
}
