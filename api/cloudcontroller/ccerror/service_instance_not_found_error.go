package ccerror

import "fmt"

// ServiceInstanceNotFoundError is returned when an endpoint cannot find the
// specified service instance
type ServiceInstanceNotFoundError struct {
	Name, SpaceGUID string
}

func (e ServiceInstanceNotFoundError) Error() string {
	return fmt.Sprintf("Service instance '%s' not found in space '%s'.", e.Name, e.SpaceGUID)
}
