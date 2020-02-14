package actionerror

import "fmt"

// SpaceSSHAlreadyDisabledError is returned when ssh is already disabled on the space
type SpaceSSHAlreadyDisabledError struct {
	Space string
	Err   error
}

func (e SpaceSSHAlreadyDisabledError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return fmt.Sprintf("ssh support for space '%s' is already disabled.", e.Space)
}
