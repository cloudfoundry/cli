package actionerror

import "fmt"

type RepositoryNameTakenError struct {
	Name string
}

func (e RepositoryNameTakenError) Error() string {
	return fmt.Sprintf("Plugin repo named '%s' already exists, please use another name.", e.Name)
}
