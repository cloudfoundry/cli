package actionerror

import "fmt"

// InvalidRouteError is returned when a route is provided that isn't properly
// formed URL.
type InvalidRouteError struct {
	Route string
}

func (err InvalidRouteError) Error() string {
	return fmt.Sprintf("Route '%s' is not parsable by URI", err.Route)
}
