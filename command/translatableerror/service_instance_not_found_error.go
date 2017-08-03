package translatableerror

type ServiceInstanceNotFoundError struct {
	GUID string
	Name string
}

func (e ServiceInstanceNotFoundError) Error() string {
	if e.Name == "" {
		return "Service instance (GUID: {{.GUID}}) not found"
	}
	return "Service instance {{.ServiceInstance}} not found"
}

func (e ServiceInstanceNotFoundError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"GUID":            e.GUID,
		"ServiceInstance": e.Name,
	})
}
