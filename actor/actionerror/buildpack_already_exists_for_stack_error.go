package actionerror

import "fmt"

type BuildpackAlreadyExistsForStackError struct {
	BuildpackName string
	StackName     string
}

func (e BuildpackAlreadyExistsForStackError) Error() string {
	return fmt.Sprintf("The buildpack name %s is already in use with stack %s", e.BuildpackName, e.StackName)
}
