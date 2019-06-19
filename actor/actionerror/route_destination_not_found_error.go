package actionerror

import "fmt"

// RouteDestinationNotFoundError is returned when a route destination cannot be found
type RouteDestinationNotFoundError struct {
	AppGUID     string
	ProcessType string
	RouteGUID   string
}

func (e RouteDestinationNotFoundError) Error() string {
	return fmt.Sprintf(
		"Destination with app guid '%s' and process type '%s' for route with guid '%s' not found.",
		e.AppGUID,
		e.ProcessType,
		e.RouteGUID,
	)
}
