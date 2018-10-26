package translatableerror

type RouterGroupNotFoundError struct {
	Name string
}

func (e RouterGroupNotFoundError) Error() string {
	return "Router group {{.RouterGroupName}} not found"
}

func (e RouterGroupNotFoundError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"RouterGroupName": e.Name,
	})
}
