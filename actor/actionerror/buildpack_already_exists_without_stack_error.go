package actionerror

import "fmt"

type BuildpackAlreadyExistsWithoutStackError struct {
	BuildpackName string
}

func (e BuildpackAlreadyExistsWithoutStackError) Error() string {
	return fmt.Sprintf("Buildpack %s already exists without a stack", e.BuildpackName)
}
