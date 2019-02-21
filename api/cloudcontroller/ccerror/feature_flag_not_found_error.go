package ccerror

// FeatureFlagNotFoundError is returned when the API endpoint is not found.
type FeatureFlagNotFoundError struct {
}

func (e FeatureFlagNotFoundError) Error() string {
	return "Feature flag not found."
}
