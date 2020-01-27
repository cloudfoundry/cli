package actionerror

import "fmt"

type SpaceQuotaAlreadyExistsError struct{ SpaceQuota string }

func (e SpaceQuotaAlreadyExistsError) Error() string {
	return fmt.Sprintf("Space quota '%s' already exists.", e.SpaceQuota)
}
