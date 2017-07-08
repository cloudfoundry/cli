package translatableerror

// RepositoryNameTakenError is returned when adding a plugin repository
// fails due to a repository already existing with the same name
type RepositoryNameTakenError struct {
	Name string
}

func (RepositoryNameTakenError) Error() string {
	return "Plugin repo named '{{.RepositoryName}}' already exists, please use another name."
}

func (e RepositoryNameTakenError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{"RepositoryName": e.Name})
}
