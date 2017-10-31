package actionerror

import "fmt"

// ServiceBindingNotFoundError is returned when a service binding cannot be
// found.
type ServiceBindingNotFoundError struct {
	AppGUID             string
	ServiceInstanceGUID string
}

func (e ServiceBindingNotFoundError) Error() string {
	return fmt.Sprintf("Service binding for application GUID '%s', and service instance GUID '%s' not found.", e.AppGUID, e.ServiceInstanceGUID)
}
