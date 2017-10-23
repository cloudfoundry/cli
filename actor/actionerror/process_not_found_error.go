package actionerror

import "fmt"

// ProcessNotFoundError is returned when a requested application is not
// found.
type ProcessNotFoundError struct {
	ProcessType string
}

func (e ProcessNotFoundError) Error() string {
	return fmt.Sprintf("Process %s not found", e.ProcessType)
}
