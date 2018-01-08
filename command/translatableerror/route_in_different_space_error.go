package translatableerror

type RouteInDifferentSpaceError struct {
	Route string
}

func (e RouteInDifferentSpaceError) Error() string {
	return "The app cannot be mapped to route {{.URL}} because the route is not in this space. Apps must be mapped to routes in the same space."
}

func (e RouteInDifferentSpaceError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"URL": e.Route,
	})
}
