package actionerror

import "fmt"

// ApplicationInstanceFlappingError is returned when an instance crashes.
type ApplicationInstanceFlappingError struct {
	Name string
}

func (e ApplicationInstanceFlappingError) Error() string {
	return fmt.Sprintf("Application '%s' crashed", e.Name)
}
