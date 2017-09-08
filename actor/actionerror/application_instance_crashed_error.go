package actionerror

import "fmt"

// ApplicationInstanceCrashedError is returned when an instance crashes.
type ApplicationInstanceCrashedError struct {
	Name string
}

func (e ApplicationInstanceCrashedError) Error() string {
	return fmt.Sprintf("Application '%s' crashed", e.Name)
}
