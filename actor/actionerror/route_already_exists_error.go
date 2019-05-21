package actionerror

import "fmt"

// RouteAlreadyExistsError is returned when a route already exists
type RouteAlreadyExistsError struct {
	Route string
	Err   error
}

func (e RouteAlreadyExistsError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return fmt.Sprintf("Route %s already exists.", e.Route)
}
