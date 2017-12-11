package translatableerror

import "fmt"

// ServiceInstanceNotSharedToSpaceError is returned when attempting to unshare a service instance from a space to which it is not shared.
type ServiceInstanceNotSharedToSpaceError struct {
	ServiceInstanceName string
}

func (e ServiceInstanceNotSharedToSpaceError) Error() string {
	return fmt.Sprintf("Failed to unshare service instance '{{.ServiceInstance}}'. Ensure the space and specified org exist and that the service instance has been shared to this space.")
}

func (e ServiceInstanceNotSharedToSpaceError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"ServiceInstance": e.ServiceInstanceName,
	})
}
