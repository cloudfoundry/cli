package actionerror

import "fmt"

// RoutePolicyAmbiguityError is returned when a route has multiple policies
// and no --source flag was given to disambiguate.
type RoutePolicyAmbiguityError struct {
	RouteURL string
	Count    int
}

func (e RoutePolicyAmbiguityError) Error() string {
	return fmt.Sprintf(
		"Route '%s' has %d policies. Specify one with --source.",
		e.RouteURL, e.Count,
	)
}
