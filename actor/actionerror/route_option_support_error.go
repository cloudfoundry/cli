package actionerror

import "fmt"

// RouteOptionSupportError is returned when route options are not supported
type RouteOptionSupportError struct {
	ErrorText string
}

func (e RouteOptionSupportError) Error() string {
	return fmt.Sprintf("Route option support: '%s'", e.ErrorText)
}
