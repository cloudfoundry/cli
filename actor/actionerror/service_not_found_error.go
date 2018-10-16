package actionerror

import "fmt"

type ServiceNotFoundError struct {
	Name string
}

func (e ServiceNotFoundError) Error() string {
	return fmt.Sprintf("Service offering '%s' not found.", e.Name)
}
