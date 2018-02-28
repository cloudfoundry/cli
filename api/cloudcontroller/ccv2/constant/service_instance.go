package constant

type ServiceInstanceType string

const (
	// UserProvidedService is a Service Instance that is created by a user.
	ServiceInstanceTypeUserProvidedService ServiceInstanceType = "user_provided_service_instance"

	// ManagedService is a Service Instance that is managed by a service broker.
	ServiceInstanceTypeManagedService ServiceInstanceType = "managed_service_instance"
)
