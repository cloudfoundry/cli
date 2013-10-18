package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"fmt"
	"strings"
)

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

type QuotaRepository interface {
	FindAll() (quotas []cf.Quota, apiResponse net.ApiResponse)
	FindByName(name string) (quota cf.Quota, apiResponse net.ApiResponse)
	Update(org cf.Organization, quota cf.Quota) (apiResponse net.ApiResponse)
}

type CloudControllerQuotaRepository struct {
	config  *configuration.Configuration
	gateway net.Gateway
}

func NewCloudControllerQuotaRepository(config *configuration.Configuration, gateway net.Gateway) (repo CloudControllerQuotaRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (repo CloudControllerQuotaRepository) findAllWithPath(path string) (quotas []cf.Quota, apiResponse net.ApiResponse) {
	resources := new(PaginatedQuotaResources)

	apiResponse = repo.gateway.GetResource(path, repo.config.AccessToken, resources)
	if apiResponse.IsNotSuccessful() {
		return
	}

	for _, r := range resources.Resources {
		quota := cf.Quota{
			Guid:        r.Metadata.Guid,
			Name:        r.Entity.Name,
			MemoryLimit: r.Entity.MemoryLimit,
		}
		quotas = append(quotas, quota)
	}

	return
}

func (repo CloudControllerQuotaRepository) FindAll() (quotas []cf.Quota, apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/quota_definitions", repo.config.Target)
	return repo.findAllWithPath(path)
}

func (repo CloudControllerQuotaRepository) FindByName(name string) (quota cf.Quota, apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/quota_definitions?q=name%%3A%s", repo.config.Target, name)
	quotas, apiResponse := repo.findAllWithPath(path)
	if apiResponse.IsNotSuccessful() {
		return
	}

	if len(quotas) == 0 {
		apiResponse = net.NewNotFoundApiResponse("Quota %s not found", name)
		return
	}

	quota = quotas[0]
	return
}

func (repo CloudControllerQuotaRepository) Update(org cf.Organization, quota cf.Quota) (apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/organizations/%s", repo.config.Target, org.Guid)
	data := fmt.Sprintf(`{"quota_definition_guid":"%s"}`, quota.Guid)
	return repo.gateway.UpdateResource(path, repo.config.AccessToken, strings.NewReader(data))
}
