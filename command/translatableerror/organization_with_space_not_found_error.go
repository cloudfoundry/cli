package translatableerror

type OrganizationWithSpaceNotFoundError struct {
	GUID      string
	Name      string
	SpaceName string
}

func (OrganizationWithSpaceNotFoundError) Error() string {
	return "Organization '{{.Name}}' containing space '{{.SpaceName}}' not found."
}

func (e OrganizationWithSpaceNotFoundError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"Name":      e.Name,
		"SpaceName": e.SpaceName,
	})
}
