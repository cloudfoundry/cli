package ccerror

import "fmt"

type ServiceOfferingNotFoundError struct {
	ServiceOfferingName, ServiceBrokerName string
}

func (e ServiceOfferingNotFoundError) Error() string {
	if e.ServiceOfferingName != "" && e.ServiceBrokerName != "" {
		return fmt.Sprintf("Service offering '%s' for service broker '%s' not found.", e.ServiceOfferingName, e.ServiceBrokerName)
	}

	if e.ServiceOfferingName != "" {
		return fmt.Sprintf("Service offering '%s' not found.", e.ServiceOfferingName)
	}

	if e.ServiceBrokerName != "" {
		return fmt.Sprintf("No service offerings found for service broker '%s'.", e.ServiceBrokerName)
	}

	return "No service offerings found."
}
