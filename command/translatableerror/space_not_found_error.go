package translatableerror

type SpaceNotFoundError struct {
	Name string
}

func (SpaceNotFoundError) Error() string {
	return "Space '{{.Name}}' not found."
}

func (e SpaceNotFoundError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"Name": e.Name,
	})
}
