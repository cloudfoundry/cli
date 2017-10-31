package actionerror

import "fmt"

// TaskNotFoundError is returned when no tasks matching the filters are found.
type TaskNotFoundError struct {
	SequenceID int
}

func (e TaskNotFoundError) Error() string {
	return fmt.Sprintf("Task sequence ID %d not found.", e.SequenceID)
}
