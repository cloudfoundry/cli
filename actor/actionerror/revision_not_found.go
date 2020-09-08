package actionerror

import "fmt"

// RevisionNotFoundError is returned when a requested application is not
// found.
type RevisionNotFoundError struct {
	Version int
	App     string
}

func (e RevisionNotFoundError) Error() string {
	return fmt.Sprintf("Revision '%d' for app '%s' not found", e.Version, e.App)
}
