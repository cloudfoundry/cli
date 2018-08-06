package actionerror

import "fmt"

// BuildpackNotFoundError is returned when a requested buildpack is not found.
type BuildpackNotFoundError struct {
	BuildpackName string
}

func (e BuildpackNotFoundError) Error() string {
	return fmt.Sprintf("Buildpack %s not found.", e.BuildpackName)
}
