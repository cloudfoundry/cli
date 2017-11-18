package translatableerror

type InvalidRouteError struct {
	Route string
}

func (InvalidRouteError) Error() string {
	return "The route '{{.Route}}' is not a properly formed URL"
}

func (e InvalidRouteError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{"Route": e.Route})
}
