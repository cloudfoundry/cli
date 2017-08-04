package translatableerror

type ApplicationAlreadyExistsError struct {
	Name string
}

func (ApplicationAlreadyExistsError) Error() string {
	return "Application '{{.Name}}' already exists."
}

func (e ApplicationAlreadyExistsError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"Name": e.Name,
	})
}
