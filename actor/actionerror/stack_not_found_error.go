package actionerror

import "fmt"

// StackNotFoundError is returned when a requested stack is not found.
type StackNotFoundError struct {
	GUID string
	Name string
}

func (e StackNotFoundError) Error() string {
	if e.Name == "" {
		return fmt.Sprintf("Stack with GUID '%s' not found.", e.GUID)
	}

	return fmt.Sprintf("Stack '%s' not found.", e.Name)
}
