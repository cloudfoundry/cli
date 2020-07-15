package translatableerror

type ServicePlanNotFoundError struct {
	PlanName          string
	OfferingName      string
	ServiceBrokerName string
}

func (e ServicePlanNotFoundError) Error() string {
	if e.OfferingName == "" {
		return "Service plan '{{.PlanName}}' not found."
	}
	if e.OfferingName != "" && e.PlanName != "" && e.ServiceBrokerName != "" {
		return "The plan '{{.PlanName}}' could not be found for service offering '{{.OfferingName}}' and broker '{{.ServiceBrokerName}}'."
	}

	return "The plan '{{.PlanName}}' could not be found for service offering '{{.OfferingName}}'."
}

func (e ServicePlanNotFoundError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"PlanName":          e.PlanName,
		"OfferingName":      e.OfferingName,
		"ServiceBrokerName": e.ServiceBrokerName,
	})
}
