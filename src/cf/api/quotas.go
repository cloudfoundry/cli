package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"fmt"
	"strings"
)

type QuotaRepository interface {
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

func (repo CloudControllerQuotaRepository) FindByName(name string) (quota cf.Quota, apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/quota_definitions?q=name%%3A%s", repo.config.Target, name)

	request, apiResponse := repo.gateway.NewRequest("GET", path, repo.config.AccessToken, nil)
	if apiResponse.IsNotSuccessful() {
		return
	}

	resources := new(PaginatedResources)

	_, apiResponse = repo.gateway.PerformRequestForJSONResponse(request, resources)
	if apiResponse.IsNotSuccessful() {
		return
	}

	if len(resources.Resources) == 0 {
		apiResponse = net.NewNotFoundApiResponse("%s %s not found", "Org", name)
		return
	}

	res := resources.Resources[0]
	quota.Guid = res.Metadata.Guid
	quota.Name = res.Entity.Name

	return
}

func (repo CloudControllerQuotaRepository) Update(org cf.Organization, quota cf.Quota) (apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/organizations/%s", repo.config.Target, org.Guid)
	data := fmt.Sprintf(`{"quota_definition_guid":"%s"}`, quota.Guid)
	request, apiResponse := repo.gateway.NewRequest("PUT", path, repo.config.AccessToken, strings.NewReader(data))
	if apiResponse.IsNotSuccessful() {
		return
	}

	apiResponse = repo.gateway.PerformRequest(request)
	return
}
