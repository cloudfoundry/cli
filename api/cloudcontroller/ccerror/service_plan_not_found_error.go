package ccerror

type ServicePlanNotFound struct{}

func (ServicePlanNotFound) Error() string {
	return "The service plan was not found"
}
