package actionerror

import "fmt"

// StartupTimeoutError is returned when startup timeout is reached waiting for
// an application to start.
type StartupTimeoutError struct {
	Name string
}

func (e StartupTimeoutError) Error() string {
	return fmt.Sprintf("Timed out waiting for application '%s' to start", e.Name)
}
