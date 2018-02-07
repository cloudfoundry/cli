package translatableerror

// ServiceInstanceNotShareableError is returned when either the
// service_instance_sharing feature flag is disabled or the service broker has
// disabled sharing
type ServiceInstanceNotShareableError struct {
	FeatureFlagEnabled          bool
	ServiceBrokerSharingEnabled bool
}

func (e ServiceInstanceNotShareableError) Error() string {
	switch {
	case !e.FeatureFlagEnabled && !e.ServiceBrokerSharingEnabled:
		return `The "service_instance_sharing" feature flag is disabled for this Cloud Foundry platform. Also, service instance sharing is disabled for this service.`
	case !e.FeatureFlagEnabled && e.ServiceBrokerSharingEnabled:
		return `The "service_instance_sharing" feature flag is disabled for this Cloud Foundry platform.`
	case e.FeatureFlagEnabled && !e.ServiceBrokerSharingEnabled:
		return "Service instance sharing is disabled for this service."
	}
	return "Unexpected ServiceInstanceNotShareableError: service instance is shareable."
}

func (e ServiceInstanceNotShareableError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error())
}
