package translatableerror

type IsolationSegmentNotFoundError struct {
	Name string
}

func (IsolationSegmentNotFoundError) Error() string {
	return "Isolation segment '{{.Name}}' not found."
}

func (e IsolationSegmentNotFoundError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"Name": e.Name,
	})
}
