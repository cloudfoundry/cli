package actionerror

import "fmt"

type RepositoryNotRegisteredError struct {
	Name string
}

func (e RepositoryNotRegisteredError) Error() string {
	return fmt.Sprintf("Plugin repository %s not found", e.Name)
}
