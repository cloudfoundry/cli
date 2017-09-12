package translatableerror

type NoMatchingDomainError struct {
	Route string
}

func (e NoMatchingDomainError) Error() string {
	return "The route {{.RouteName}} did not match any existing domains."
}

func (e NoMatchingDomainError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"RouteName": e.Route,
	})
}
