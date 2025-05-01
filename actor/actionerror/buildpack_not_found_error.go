package actionerror

import "fmt"

// BuildpackNotFoundError is returned when a requested buildpack is not found.
type BuildpackNotFoundError struct {
	BuildpackName string
	StackName     string
	Lifecycle     string
}

func (e BuildpackNotFoundError) Error() string {
	return fmt.Sprintf("Buildpack not found - Name: '%s'; Stack: '%s', Lifecyle: '%s'", e.BuildpackName, e.StackName, e.Lifecycle)
}
