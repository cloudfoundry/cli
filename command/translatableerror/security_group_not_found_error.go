package translatableerror

type SecurityGroupNotFoundError struct {
	Name string
}

func (SecurityGroupNotFoundError) Error() string {
	return "Security group '{{.Name}}' not found."
}

func (e SecurityGroupNotFoundError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"Name": e.Name,
	})
}
