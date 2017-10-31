package actionerror

import "fmt"

type ServiceInstanceNotFoundError struct {
	GUID string
	Name string
}

func (e ServiceInstanceNotFoundError) Error() string {
	if e.Name == "" {
		return fmt.Sprintf("Service instance (GUID: %s) not found.", e.GUID)
	}
	return fmt.Sprintf("Service instance '%s' not found.", e.Name)
}
