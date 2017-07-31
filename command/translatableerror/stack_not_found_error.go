package translatableerror

type StackNotFoundError struct {
	GUID string
	Name string
}

func (e StackNotFoundError) Error() string {
	if e.Name == "" {
		return "Stack with GUID {{.GUID}} not found"
	}

	return "Stack {{.Name}} not found"
}

func (e StackNotFoundError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"GUID": e.GUID,
		"Name": e.Name,
	})
}
