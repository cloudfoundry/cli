package actionerror

type RouteBindingNotFoundError struct{}

func (e RouteBindingNotFoundError) Error() string {
	return "Route binding not found."
}
