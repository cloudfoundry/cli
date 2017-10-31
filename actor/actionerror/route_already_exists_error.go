package actionerror

import "fmt"

// RouteAlreadyExistsError is returned when a route already exists
type RouteAlreadyExistsError struct {
	Route string
}

func (e RouteAlreadyExistsError) Error() string {
	return fmt.Sprintf("Route %s already exists", e.Route)
}
