package actionerror

import "fmt"

type SpaceQuotaNotFoundError struct {
	GUID string
}

func (e SpaceQuotaNotFoundError) Error() string {
	return fmt.Sprintf("Space quota with GUID '%s' not found.", e.GUID)
}
