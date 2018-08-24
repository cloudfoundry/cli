package actionerror

import "fmt"

// EmptyBuildpackDirectoryError represents the error when a user tries to upload
// new buildpack bits but specifies an empty directory as the path to these bits.
type EmptyBuildpackDirectoryError struct {
	Path string
}

func (e EmptyBuildpackDirectoryError) Error() string {
	return fmt.Sprint(e.Path, " is empty")
}
