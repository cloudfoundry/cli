package actionerror

import "fmt"

// IsolationSegmentNotFoundError represents the error that occurs when the
// isolation segment is not found.
type IsolationSegmentNotFoundError struct {
	Name string
}

func (e IsolationSegmentNotFoundError) Error() string {
	return fmt.Sprintf("Isolation Segment '%s' not found.", e.Name)
}
