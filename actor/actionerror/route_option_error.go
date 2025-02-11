package actionerror

import "fmt"

// RouteOptionError is returned when a route option was specified in the wrong format
type RouteOptionError struct {
	Name       string
	Host       string
	DomainName string
	Path       string
}

func (e RouteOptionError) Error() string {
	return fmt.Sprintf("Route option '%s' for route with host '%s', domain '%s', and path '%s' was specified incorrectly. Please use key-value pair format key=value.", e.Name, e.Host, e.DomainName, e.path())
}

func (e RouteOptionError) path() string {
	if e.Path == "" {
		return "/"
	}
	return e.Path
}
