package actionerror

import "fmt"

type AddPluginRepositoryError struct {
	Name    string
	URL     string
	Message string
}

func (e AddPluginRepositoryError) Error() string {
	return fmt.Sprintf("Could not add repository '%s' from %s: %s", e.Name, e.URL, e.Message)
}
