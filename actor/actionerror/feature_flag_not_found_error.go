package actionerror

import "fmt"

// FeatureFlagNotFoundError is returned when a requested feature flag is not found.
type FeatureFlagNotFoundError struct {
	FeatureFlagName string
}

func (e FeatureFlagNotFoundError) Error() string {
	return fmt.Sprintf("Feature flag '%s' not found.", e.FeatureFlagName)
}
