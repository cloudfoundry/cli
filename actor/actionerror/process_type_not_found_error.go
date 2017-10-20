package actionerror

import "fmt"

// ProcessTypeNotFoundError is returned when a requested application is not
// found.
type ProcessTypeNotFoundError struct {
	ProcessType string
}

func (e ProcessTypeNotFoundError) Error() string {
	return fmt.Sprintf("Process %s not found", e.ProcessType)
}
