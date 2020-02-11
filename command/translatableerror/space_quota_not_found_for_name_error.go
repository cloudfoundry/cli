package translatableerror

type SpaceQuotaNotFoundForNameError struct {
	Name string
}

func (SpaceQuotaNotFoundForNameError) Error() string {
	return "Space quota with name '{{.Name}}' not found."
}

func (e SpaceQuotaNotFoundForNameError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"Name": e.Name,
	})
}
