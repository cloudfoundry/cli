package actionerror

// RouteInDifferentSpaceError is returned when the route exists in a different
// space than the one requesting it.
type RouteInDifferentSpaceError struct {
	Route string
}

func (RouteInDifferentSpaceError) Error() string {
	return "route registered to another space"
}
