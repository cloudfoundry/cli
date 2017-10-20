package v2action

type ServiceInstanceSummary struct {
	ServiceInstance

	ServicePlan       ServicePlan
	Service           Service
	BoundApplications []string
}

func (actor Actor) GetServiceInstanceSummaryByNameAndSpace(name string, spaceGUID string) (ServiceInstanceSummary, Warnings, error) {
	serviceInstance, warnings, err := actor.GetServiceInstanceByNameAndSpace(name, spaceGUID)
	allWarnings := Warnings(warnings)
	if err != nil {
		return ServiceInstanceSummary{}, allWarnings, err
	}

	servicePlan, warnings, err := actor.GetServicePlan(serviceInstance.ServicePlanGUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return ServiceInstanceSummary{}, allWarnings, err
	}

	service, warnings, err := actor.GetService(servicePlan.ServiceGUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return ServiceInstanceSummary{}, allWarnings, err
	}

	serviceBindings, warnings, err := actor.GetServiceBindingsByServiceInstance(serviceInstance.GUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return ServiceInstanceSummary{}, allWarnings, err
	}

	boundApps := []string{}
	for _, serviceBinding := range serviceBindings {
		app, warnings, err := actor.GetApplication(serviceBinding.AppGUID)
		allWarnings = append(allWarnings, warnings...)
		if err != nil {
			return ServiceInstanceSummary{}, allWarnings, err
		}
		boundApps = append(boundApps, app.Name)
	}

	return ServiceInstanceSummary{
		ServiceInstance:   serviceInstance,
		ServicePlan:       servicePlan,
		Service:           service,
		BoundApplications: boundApps,
	}, allWarnings, nil
}
