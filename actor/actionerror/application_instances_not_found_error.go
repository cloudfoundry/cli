package actionerror

import "fmt"

// ApplicationInstancesNotFoundError is returned when the application does not
// have running instances.
type ApplicationInstancesNotFoundError struct {
	ApplicationGUID string
}

func (e ApplicationInstancesNotFoundError) Error() string {
	return fmt.Sprintf("Application instances '%s' not found.", e.ApplicationGUID)
}
