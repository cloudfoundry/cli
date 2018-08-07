package actionerror

import "fmt"

type BuildpackNameTakenError struct {
	Name string
}

func (e BuildpackNameTakenError) Error() string {
	return fmt.Sprintf("Buildpack %s already exists", e.Name)
}
