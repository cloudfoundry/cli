package constant

// FeatureFlagName is the name of the feature flag given by the Cloud
// Controller.
type FeatureFlagName string

const (
	// FeatureFlagServiceInstanceSharing is the name of the service instance
	// sharing feature flag.
	FeatureFlagServiceInstanceSharing FeatureFlagName = "service_instance_sharing"
)
