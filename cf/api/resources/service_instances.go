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

type ServiceInstanceEntity struct {
	Name             string
	DashboardUrl     string                   `json:"dashboard_url"`
	ServiceBindings  []ServiceBindingResource `json:"service_bindings"`
	ServicePlan      ServicePlanResource      `json:"service_plan"`
	State            string                   `json:"state"`
	StateDescription string                   `json:"state_description"`
}

func (resource ServiceInstanceResource) ToFields() (fields models.ServiceInstanceFields) {
	fields.Guid = resource.Metadata.Guid
	fields.Name = resource.Entity.Name
	fields.State = resource.Entity.State
	fields.StateDescription = resource.Entity.StateDescription
	fields.DashboardUrl = resource.Entity.DashboardUrl
	return
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
