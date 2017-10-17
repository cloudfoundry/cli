package actionerror

import "fmt"

// ProcessTypeNotFoundError is returned when a requested application is not
// found.
type ProcessTypeNotFoundError struct {
	Name string
}

func (e ProcessTypeNotFoundError) Error() string {
	return fmt.Sprintf("Process type '%s' not found.", e.Name)
}
