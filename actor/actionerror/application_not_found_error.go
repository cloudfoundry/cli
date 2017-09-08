package actionerror

import "fmt"

// ApplicationNotFoundError is returned when a requested application is not
// found.
type ApplicationNotFoundError struct {
	GUID string
	Name string
}

func (e ApplicationNotFoundError) Error() string {
	if e.GUID != "" {
		return fmt.Sprintf("Application with GUID '%s' not found.", e.GUID)
	}

	return fmt.Sprintf("Application '%s' not found.", e.Name)
}
