package actionerror

import "fmt"

type SpaceNameTakenError struct {
	Name string
}

func (err SpaceNameTakenError) Error() string {
	return fmt.Sprintf("A space with the name '%s' already exists", err.Name)
}
