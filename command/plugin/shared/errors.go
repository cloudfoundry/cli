package shared

type PluginNotFoundError struct {
	Name string
}

func (e PluginNotFoundError) Error() string {
	return "Plugin {{.Name}} does not exist."
}

func (e PluginNotFoundError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"Name": e.Name,
	})
}

type NoPluginRepositoriesError struct{}

func (e NoPluginRepositoriesError) Error() string {
	return "No plugin repositories registered to search for plugin updates."
}

func (e NoPluginRepositoriesError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error())
}

// GettingPluginRepositoryError is returned when there's an error
// accessing the plugin repository
type GettingPluginRepositoryError struct {
	Name    string
	Message string
}

func (e GettingPluginRepositoryError) Error() string {
	return "Could not get plugin repository '{{.RepositoryName}}': {{.ErrorMessage}}"
}

func (e GettingPluginRepositoryError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{"RepositoryName": e.Name, "ErrorMessage": e.Message})
}

// RepositoryNameTakenError is returned when adding a plugin repository
// fails due to a repository already existing with the same name
type RepositoryNameTakenError struct {
	Name string
}

func (e RepositoryNameTakenError) Error() string {
	return "Plugin repo named '{{.RepositoryName}}' already exists, please use another name."
}

func (e RepositoryNameTakenError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{"RepositoryName": e.Name})
}

// RepositoryURLTakenError is returned when adding a plugin repository
// fails due to a repository already existing with the same URL
type RepositoryURLTakenError struct {
	Name string
	URL  string
}

func (e RepositoryURLTakenError) Error() string {
	return "{{.RepositoryURL}} ({{.RepositoryName}}) already exists."
}

func (e RepositoryURLTakenError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"RepositoryName": e.Name,
		"RepositoryURL":  e.URL,
	})
}

type AddPluginRepositoryError struct {
	Name    string
	URL     string
	Message string
}

func (e AddPluginRepositoryError) Error() string {
	return "Could not add repository '{{.RepositoryName}}' from {{.RepositoryURL}}: {{.Message}}"
}

func (e AddPluginRepositoryError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"RepositoryName": e.Name,
		"RepositoryURL":  e.URL,
		"Message":        e.Message,
	})
}
