package actionerror

import "fmt"

// RevisionNotFoundError is returned when a requested application is not
// found.
type RevisionNotFoundError struct {
	Version int
}

func (e RevisionNotFoundError) Error() string {
	return fmt.Sprintf("Revision (%d) not found", e.Version)
}

type RevisionAmbiguousError struct {
	Version int
}

func (e RevisionAmbiguousError) Error() string {
	return fmt.Sprintf("More than one revision (%d) found", e.Version)
}
