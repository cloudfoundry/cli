package actionerror

import "fmt"

type ServicePlanNotFoundError struct {
	PlanName    string
	ServiceName string
}

func (e ServicePlanNotFoundError) Error() string {
	if e.ServiceName == "" {
		return fmt.Sprintf("Service plan '%s' not found.", e.PlanName)
	}
	return fmt.Sprintf("The plan %s could not be found for service %s", e.PlanName, e.ServiceName)
}
