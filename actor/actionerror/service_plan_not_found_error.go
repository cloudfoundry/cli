package actionerror

import "fmt"

type ServicePlanNotFoundError struct {
	Name        string
	ServiceName string
}

func (e ServicePlanNotFoundError) Error() string {
	return fmt.Sprintf("The plan '%s' could not be found for service %s", e.Name, e.ServiceName)
}
