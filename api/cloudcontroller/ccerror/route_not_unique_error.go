package ccerror

type RouteNotUniqueError struct {
}

func (e RouteNotUniqueError) Error() string {
	return "Route already exists for domain"
}
