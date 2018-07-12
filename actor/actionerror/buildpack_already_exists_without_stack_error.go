package actionerror

import "fmt"

type BuildpackAlreadyExistsWithoutStackError string

func (e BuildpackAlreadyExistsWithoutStackError) Error() string {
	return fmt.Sprintf("Buildpack %s already exists without a stack", string(e))
}
