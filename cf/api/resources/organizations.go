package resources

import "github.com/cloudfoundry/cli/cf/models"

type OrganizationResource struct {
	Resource
	Entity OrganizationEntity
}

type OrganizationEntity struct {
	Name            string
	QuotaDefinition QuotaResource `json:"quota_definition"`
	Spaces          []SpaceResource
	Domains         []DomainResource
}

func (resource OrganizationResource) ToFields() (fields models.OrganizationFields) {
	fields.Name = resource.Entity.Name
	fields.Guid = resource.Metadata.Guid

	fields.QuotaDefinition = resource.Entity.QuotaDefinition.ToFields()
	return
}

func (resource OrganizationResource) ToModel() (org models.Organization) {
	org.OrganizationFields = resource.ToFields()

	spaces := []models.SpaceFields{}
	for _, s := range resource.Entity.Spaces {
		spaces = append(spaces, s.ToFields())
	}
	org.Spaces = spaces

	domains := []models.DomainFields{}
	for _, d := range resource.Entity.Domains {
		domains = append(domains, d.ToFields())
	}
	org.Domains = domains

	return
}
