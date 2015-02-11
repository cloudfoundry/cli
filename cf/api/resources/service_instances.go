package resources

import "github.com/cloudfoundry/cli/cf/models"

type PaginatedServiceInstanceResources struct {
	TotalResults int `json:"total_results"`
	Resources    []ServiceInstanceResource
}

type ServiceInstanceResource struct {
	Resource
	Entity ServiceInstanceEntity
}

type LastOperation struct {
	Type        string `json:"type"`
	State       string `json:"state"`
	Description string `json:"description"`
}

type ServiceInstanceEntity struct {
	Name            string
	DashboardUrl    string                   `json:"dashboard_url"`
	ServiceBindings []ServiceBindingResource `json:"service_bindings"`
	ServicePlan     ServicePlanResource      `json:"service_plan"`
	LastOperation   LastOperation            `json:"last_operation"`
}

func (resource ServiceInstanceResource) ToFields() models.ServiceInstanceFields {
	return models.ServiceInstanceFields{
		Guid:         resource.Metadata.Guid,
		Name:         resource.Entity.Name,
		DashboardUrl: resource.Entity.DashboardUrl,
		LastOperation: models.LastOperationFields{
			Type:        resource.Entity.LastOperation.Type,
			State:       resource.Entity.LastOperation.State,
			Description: resource.Entity.LastOperation.Description,
		},
	}
}

func (resource ServiceInstanceResource) ToModel() (instance models.ServiceInstance) {
	instance.ServiceInstanceFields = resource.ToFields()
	instance.ServicePlan = resource.Entity.ServicePlan.ToFields()

	instance.ServiceBindings = []models.ServiceBindingFields{}
	for _, bindingResource := range resource.Entity.ServiceBindings {
		instance.ServiceBindings = append(instance.ServiceBindings, bindingResource.ToFields())
	}
	return
}
