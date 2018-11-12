package actionerror

import "fmt"

type BuildpackStackChangeError struct {
	BuildpackName string
}

func (e BuildpackStackChangeError) Error() string {
	return fmt.Sprintf("Buildpack %s already exists with a stack association", e.BuildpackName)
}
