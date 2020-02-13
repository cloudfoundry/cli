package actionerror

import "fmt"

// SpaceSSHAlreadyEnabledError is returned when ssh is already enabled on the space
type SpaceSSHAlreadyEnabledError struct {
	Space string
	Err   error
}

func (e SpaceSSHAlreadyEnabledError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return fmt.Sprintf("ssh support for space '%s' is already enabled.", e.Space)
}
