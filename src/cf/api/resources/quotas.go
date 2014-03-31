package resources

import "cf/models"

type PaginatedQuotaResources struct {
	Resources []QuotaResource
}

type QuotaResource struct {
	Resource
	Entity QuotaEntity
}

type QuotaEntity struct {
	Name        string
	MemoryLimit uint64 `json:"memory_limit"`
}

func (resource QuotaResource) ToFields() (quota models.QuotaFields) {
	quota.Guid = resource.Metadata.Guid
	quota.Name = resource.Entity.Name
	quota.MemoryLimit = resource.Entity.MemoryLimit
	return
}
