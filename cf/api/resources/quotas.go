package resources

import "github.com/cloudfoundry/cli/cf/models"

type PaginatedQuotaResources struct {
	Resources []QuotaResource
}

type QuotaResource struct {
	Resource
	Entity models.QuotaResponse
}

func (resource QuotaResource) ToFields() models.QuotaFields {
	var appInstanceLimit int = UnlimitedAppInstances
	if resource.Entity.AppInstanceLimit != "" {
		i, err := resource.Entity.AppInstanceLimit.Int64()
		if err == nil {
			appInstanceLimit = int(i)
		}
	}

	return models.QuotaFields{
		Guid:                    resource.Metadata.Guid,
		Name:                    resource.Entity.Name,
		MemoryLimit:             resource.Entity.MemoryLimit,
		InstanceMemoryLimit:     resource.Entity.InstanceMemoryLimit,
		RoutesLimit:             resource.Entity.RoutesLimit,
		ServicesLimit:           resource.Entity.ServicesLimit,
		NonBasicServicesAllowed: resource.Entity.NonBasicServicesAllowed,
		AppInstanceLimit:        appInstanceLimit,
	}
}
