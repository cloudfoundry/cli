package actionerror

import "fmt"

type ServicePlanNotFoundError struct {
	PlanName          string
	OfferingName      string
	ServiceBrokerName string
}

func (e ServicePlanNotFoundError) Error() string {
	if e.OfferingName != "" && e.PlanName != "" {
		if e.ServiceBrokerName != "" {
			return fmt.Sprintf("The plan %s could not be found for service offering %s and broker %s.", e.PlanName, e.OfferingName, e.ServiceBrokerName)
		}
		return fmt.Sprintf("The plan %s could not be found for service offering %s.", e.PlanName, e.OfferingName)
	}

	if e.PlanName != "" {
		return fmt.Sprintf("Service plan '%s' not found.", e.PlanName)
	}

	if e.OfferingName != "" {
		return fmt.Sprintf("No service plans found for service offering '%s'.", e.OfferingName)
	}

	return "No service plans found."
}
