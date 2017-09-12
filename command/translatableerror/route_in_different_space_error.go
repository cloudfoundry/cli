package translatableerror

type RouteInDifferentSpaceError struct {
	Route string
}

func (e RouteInDifferentSpaceError) Error() string {
	return "The route {{.URL}} is already in use.\nTIP: Change the hostname with -n HOSTNAME or use --random-route to generate a new route and then push again."
}

func (e RouteInDifferentSpaceError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"URL": e.Route,
	})
}
