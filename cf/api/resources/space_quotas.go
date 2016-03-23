package resources

import "github.com/cloudfoundry/cli/cf/models"

type PaginatedSpaceQuotaResources struct {
	Resources []SpaceQuotaResource
}

type SpaceQuotaResource struct {
	Resource
	Entity models.SpaceQuotaResponse
}

func (resource SpaceQuotaResource) ToModel() models.SpaceQuota {
	entity := resource.Entity

	appInstanceLimit := UnlimitedAppInstances
	if entity.AppInstanceLimit != "" {
		i, err := entity.AppInstanceLimit.Int64()
		if err == nil {
			appInstanceLimit = int(i)
		}
	}

	return models.SpaceQuota{
		Guid:                    resource.Metadata.Guid,
		Name:                    entity.Name,
		MemoryLimit:             entity.MemoryLimit,
		InstanceMemoryLimit:     entity.InstanceMemoryLimit,
		RoutesLimit:             entity.RoutesLimit,
		ServicesLimit:           entity.ServicesLimit,
		NonBasicServicesAllowed: entity.NonBasicServicesAllowed,
		OrgGuid:                 entity.OrgGuid,
		AppInstanceLimit:        appInstanceLimit,
	}
}
