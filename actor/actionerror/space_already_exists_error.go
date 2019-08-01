package actionerror

import "fmt"

// SpaceAlreadyExistsError is returned when a space already exists
type SpaceAlreadyExistsError struct {
	Space string
	Err   error
}

func (e SpaceAlreadyExistsError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return fmt.Sprintf("Space '%s' already exists.", e.Space)
}
