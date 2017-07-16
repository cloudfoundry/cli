package translatableerror

type RouteInDifferentSpaceError struct {
	Route string
}

func (e RouteInDifferentSpaceError) Error() string {
	return "Route {{.Route}} has been registered to another space."
}

func (e RouteInDifferentSpaceError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"Route": e.Route,
	})
}
