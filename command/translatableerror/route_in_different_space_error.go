package translatableerror

type RouteInDifferentSpaceError struct {
	Route string
}

func (e RouteInDifferentSpaceError) Error() string {
	return "The app cannot be mapped to route {{.URL}} because the route exists in a different space."
}

func (e RouteInDifferentSpaceError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"URL": e.Route,
	})
}
