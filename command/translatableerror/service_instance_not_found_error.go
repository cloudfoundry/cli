package translatableerror

type ServiceInstanceNotFoundError struct {
	Name string
}

func (_ ServiceInstanceNotFoundError) Error() string {
	return "Service instance {{.ServiceInstance}} not found"
}

func (e ServiceInstanceNotFoundError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"ServiceInstance": e.Name,
	})
}
