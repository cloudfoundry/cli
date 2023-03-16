package actionerror

import "fmt"

// RevisionAmbiguousError is returned when multiple revisions with the same
// version are returned
type RevisionAmbiguousError struct {
	Version int
}

func (e RevisionAmbiguousError) Error() string {
	return fmt.Sprintf("More than one revision '%d' found", e.Version)
}
