package actionerror

import "fmt"

// IsolationSegmentAlreadyExistsError gets returned when an isolation segment
// already exists.
type IsolationSegmentAlreadyExistsError struct {
	Name string
}

func (e IsolationSegmentAlreadyExistsError) Error() string {
	return fmt.Sprintf("Isolation Segment '%s' already exists.", e.Name)
}
