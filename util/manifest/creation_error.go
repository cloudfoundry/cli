package manifest

import "fmt"

type ManifestCreationError struct {
	Err error
}

func (e ManifestCreationError) Error() string {
	return fmt.Sprintf("Error creating manifest file: %s", e.Err.Error())
}
