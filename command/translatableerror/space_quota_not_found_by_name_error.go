package translatableerror

type SpaceQuotaNotFoundByNameError struct {
	Name string
}

func (SpaceQuotaNotFoundByNameError) Error() string {
	return "Space quota with name '{{.Name}}' not found."
}

func (e SpaceQuotaNotFoundByNameError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"Name": e.Name,
	})
}
