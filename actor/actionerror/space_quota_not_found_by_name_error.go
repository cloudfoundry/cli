package actionerror

import "fmt"

type SpaceQuotaNotFoundByNameError struct {
	Name string
}

//TODO: confirm error wording with Abby
func (e SpaceQuotaNotFoundByNameError) Error() string {
	return fmt.Sprintf("Space quota with name '%s' not found.", e.Name)
}
