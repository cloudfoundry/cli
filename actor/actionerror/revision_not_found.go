package actionerror

import "fmt"

// RevisionNotFoundError is returned when a requested revision is not
// found.
type RevisionNotFoundError struct {
	Version int
}

func (e RevisionNotFoundError) Error() string {
	return fmt.Sprintf("Revision '%d' not found", e.Version)
}
