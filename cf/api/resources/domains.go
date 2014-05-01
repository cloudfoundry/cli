package resources

import "github.com/cloudfoundry/cli/cf/models"

type DomainResource struct {
	Resource
	Entity DomainEntity
}

type DomainEntity struct {
	Name                   string `json:"name"`
	OwningOrganizationGuid string `json:"owning_organization_guid,omitempty"`
	Wildcard               bool   `json:"wildcard"`
}

func (resource DomainResource) ToFields() models.DomainFields {
	owningOrganizationGuid := resource.Entity.OwningOrganizationGuid
	return models.DomainFields{
		Name: resource.Entity.Name,
		Guid: resource.Metadata.Guid,
		OwningOrganizationGuid: owningOrganizationGuid,
		Shared:                 owningOrganizationGuid == "",
	}
}
