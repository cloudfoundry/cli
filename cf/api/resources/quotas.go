package resources

import "github.com/cloudfoundry/cli/cf/models"

type PaginatedQuotaResources struct {
	Resources []QuotaResource
}

type QuotaResource struct {
	Resource
	Entity models.QuotaFields
}

type QuotaUsageResource struct {
	Resource
	Entity models.QuotaUsage
}

func (resource QuotaResource) ToFields() (quota models.QuotaFields) {
	quota.Guid = resource.Metadata.Guid
	quota.Name = resource.Entity.Name
	quota.MemoryLimit = resource.Entity.MemoryLimit
	quota.InstanceMemoryLimit = resource.Entity.InstanceMemoryLimit
	quota.RoutesLimit = resource.Entity.RoutesLimit
	quota.ServicesLimit = resource.Entity.ServicesLimit
	quota.NonBasicServicesAllowed = resource.Entity.NonBasicServicesAllowed
	return
}

func (resource QuotaUsageResource) ToFields() (quotaUsage models.QuotaUsage) {
	quotaUsage.Guid = resource.Metadata.Guid
	quotaUsage.Name = resource.Entity.Name
	quotaUsage.MemoryLimit = resource.Entity.MemoryLimit
	quotaUsage.InstanceMemoryLimit = resource.Entity.InstanceMemoryLimit
	quotaUsage.RoutesLimit = resource.Entity.RoutesLimit
	quotaUsage.ServicesLimit = resource.Entity.ServicesLimit
	quotaUsage.NonBasicServicesAllowed = resource.Entity.NonBasicServicesAllowed
	quotaUsage.OrgUsage.Routes = resource.Entity.OrgUsage.Routes
	quotaUsage.OrgUsage.Services = resource.Entity.OrgUsage.Services
	quotaUsage.OrgUsage.Memory = resource.Entity.OrgUsage.Memory
	return
}
