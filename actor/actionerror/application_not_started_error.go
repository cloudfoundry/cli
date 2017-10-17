package actionerror

import "fmt"

// ApplicationNotStartedError is returned when trying to ssh to an application that is not STARTED.
type ApplicationNotStartedError struct {
	Name string
}

func (e ApplicationNotStartedError) Error() string {
	return fmt.Sprintf("Application '%s' is not in the STARTED state", e.Name)
}
