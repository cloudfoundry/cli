package actionerror

import "fmt"

// RouteBoundToMultipleAppsError is returned when a route is mapped to more than one app
type RouteBoundToMultipleAppsError struct {
	AppName  string
	RouteURL string
}

func (e RouteBoundToMultipleAppsError) Error() string {
	return fmt.Sprintf("App '%s' was not deleted because route '%s' is mapped to more than one app.", e.AppName, e.RouteURL)
}
