package resources

import "github.com/cloudfoundry/cli/cf/models"

type SpaceResource struct {
	Resource
	Entity SpaceEntity
}

type SpaceEntity struct {
	Name             string
	Organization     OrganizationResource
	Applications     []ApplicationResource `json:"apps"`
	Domains          []DomainResource
	ServiceInstances []ServiceInstanceResource `json:"service_instances"`
}

func (resource SpaceResource) ToFields() (fields models.SpaceFields) {
	fields.Guid = resource.Metadata.Guid
	fields.Name = resource.Entity.Name
	return
}

func (resource SpaceResource) ToModel() (space models.Space) {
	space.SpaceFields = resource.ToFields()
	for _, app := range resource.Entity.Applications {
		space.Applications = append(space.Applications, app.ToFields())
	}

	for _, domainResource := range resource.Entity.Domains {
		space.Domains = append(space.Domains, domainResource.ToFields())
	}

	for _, serviceResource := range resource.Entity.ServiceInstances {
		space.ServiceInstances = append(space.ServiceInstances, serviceResource.ToFields())
	}

	space.Organization = resource.Entity.Organization.ToFields()
	return
}
