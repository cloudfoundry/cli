package translatableerror

type RepositoryNotRegisteredError struct {
	Name string
}

func (RepositoryNotRegisteredError) Error() string {
	return "Plugin repository {{.Name}} not found.\nUse 'cf list-plugin-repos' to list registered repos."
}

func (e RepositoryNotRegisteredError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"Name": e.Name,
	})
}
