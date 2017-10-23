package actionerror

import "fmt"

type RepositoryAlreadyExistsError struct {
	Name string
	URL  string
}

func (e RepositoryAlreadyExistsError) Error() string {
	return fmt.Sprintf("%s already registered as %s.", e.URL, e.Name)
}
