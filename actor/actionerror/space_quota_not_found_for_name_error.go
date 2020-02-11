package actionerror

import "fmt"

type SpaceQuotaNotFoundForNameError struct {
	Name string
}

func (e SpaceQuotaNotFoundForNameError) Error() string {
	return fmt.Sprintf("Space quota with name '%s' not found.", e.Name)
}
