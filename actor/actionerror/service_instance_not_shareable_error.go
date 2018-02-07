package actionerror

// ServiceInstanceNotShareableError is returned when either the
// service_instance_sharing feature flag is disabled or the service broker has
// disabled sharing
type ServiceInstanceNotShareableError struct {
	FeatureFlagEnabled          bool
	ServiceBrokerSharingEnabled bool
}

func (e ServiceInstanceNotShareableError) Error() string {
	return "Service instance sharing is not enabled"
}
