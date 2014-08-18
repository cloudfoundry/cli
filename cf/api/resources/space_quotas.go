package resources

import "github.com/cloudfoundry/cli/cf/models"

type PaginatedSpaceQuotaResources struct {
	Resources []SpaceQuotaResource
}

type SpaceQuotaResource struct {
	Resource
	Entity models.SpaceQuota
}

func (resource SpaceQuotaResource) ToModel() models.SpaceQuota {
	entity := resource.Entity

	return models.SpaceQuota{
		Guid:                    resource.Metadata.Guid,
		Name:                    entity.Name,
		MemoryLimit:             entity.MemoryLimit,
		RoutesLimit:             entity.RoutesLimit,
		ServicesLimit:           entity.ServicesLimit,
		NonBasicServicesAllowed: entity.NonBasicServicesAllowed,
		OrgGuid:                 entity.OrgGuid,
	}
}
