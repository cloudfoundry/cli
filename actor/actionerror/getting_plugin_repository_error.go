package actionerror

import "fmt"

// GettingPluginRepositoryError is returned when there's an error
// accessing the plugin repository.
type GettingPluginRepositoryError struct {
	Name    string
	Message string
}

func (e GettingPluginRepositoryError) Error() string {
	return fmt.Sprintf("Could not get plugin repository '%s'\n%s", e.Name, e.Message)
}
