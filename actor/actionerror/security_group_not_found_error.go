package actionerror

import "fmt"

// SecurityGroupNotFoundError is returned when a requested security group is
// not found.
type SecurityGroupNotFoundError struct {
	Name string
}

func (e SecurityGroupNotFoundError) Error() string {
	return fmt.Sprintf("Security group '%s' not found.", e.Name)
}
