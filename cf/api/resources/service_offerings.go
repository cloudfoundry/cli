package resources

import "github.com/cloudfoundry/cli/cf/models"

type PaginatedServiceOfferingResources struct {
	Resources []ServiceOfferingResource
}

type ServiceOfferingResource struct {
	Resource
	Entity ServiceOfferingEntity
}

type ServiceOfferingEntity struct {
	Label            string                `json:"label"`
	Version          string                `json:"version"`
	Description      string                `json:"description"`
	DocumentationUrl string                `json:"documentation_url"`
	Provider         string                `json:"provider"`
	BrokerGuid       string                `json:"service_broker_guid"`
	ServicePlans     []ServicePlanResource `json:"service_plans"`
}

func (resource ServiceOfferingResource) ToFields() (fields models.ServiceOfferingFields) {
	fields.Label = resource.Entity.Label
	fields.Version = resource.Entity.Version
	fields.Provider = resource.Entity.Provider
	fields.Description = resource.Entity.Description
	fields.BrokerGuid = resource.Entity.BrokerGuid
	fields.Guid = resource.Metadata.Guid
	fields.DocumentationUrl = resource.Entity.DocumentationUrl
	return
}

func (resource ServiceOfferingResource) ToModel() (offering models.ServiceOffering) {
	offering.ServiceOfferingFields = resource.ToFields()
	for _, p := range resource.Entity.ServicePlans {
		servicePlan := models.ServicePlanFields{}
		servicePlan.Name = p.Entity.Name
		servicePlan.Guid = p.Metadata.Guid
		offering.Plans = append(offering.Plans, servicePlan)
	}
	return offering
}
