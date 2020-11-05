package actionerror

type RouteBindingNotFoundError struct{}

func (e RouteBindingNotFoundError) Error() string {
	return "route binding not found"
}
