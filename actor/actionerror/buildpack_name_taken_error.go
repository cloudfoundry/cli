package actionerror

import "fmt"

type BuildpackNameTakenError string

func (e BuildpackNameTakenError) Error() string {
	return fmt.Sprintf("Buildpack %s already exists", string(e))
}
