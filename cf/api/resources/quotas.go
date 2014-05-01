package resources

import "github.com/cloudfoundry/cli/cf/models"

type PaginatedQuotaResources struct {
	Resources []QuotaResource
}

type QuotaResource struct {
	Resource
	Entity models.QuotaFields
}

func (resource QuotaResource) ToFields() (quota models.QuotaFields) {
	quota.Guid = resource.Metadata.Guid
	quota.Name = resource.Entity.Name
	quota.MemoryLimit = resource.Entity.MemoryLimit
	quota.RoutesLimit = resource.Entity.RoutesLimit
	quota.ServicesLimit = resource.Entity.ServicesLimit
	quota.NonBasicServicesAllowed = resource.Entity.NonBasicServicesAllowed
	return
}
