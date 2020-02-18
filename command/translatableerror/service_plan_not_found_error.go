package translatableerror

type ServicePlanNotFoundError struct {
	PlanName    string
	ServiceName string
}

func (e ServicePlanNotFoundError) Error() string {
	if e.ServiceName == "" {
		return "Service plan '{{.PlanName}}' not found."
	}
	return "The plan {{.PlanName}} could not be found for service {{.ServiceName}}"
}

func (e ServicePlanNotFoundError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"PlanName":    e.PlanName,
		"ServiceName": e.ServiceName,
	})
}
