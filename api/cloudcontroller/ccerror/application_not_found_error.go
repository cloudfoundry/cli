package ccerror

import "fmt"

// ApplicationNotFoundError is returned when an endpoint cannot find the
// specified application
type ApplicationNotFoundError struct {
	Name string
}

func (e ApplicationNotFoundError) Error() string {
	if e.Name != "" {
		return fmt.Sprintf("App '%s' not found.", e.Name)
	}

	return "Application not found"
}
