package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"fmt"
	"strings"
)

type OrganizationRepository interface {
	FindAll() (orgs []cf.Organization, apiResponse net.ApiResponse)
	FindByName(name string) (org cf.Organization, apiResponse net.ApiResponse)
	Create(name string) (apiResponse net.ApiResponse)
	Rename(org cf.Organization, name string) (apiResponse net.ApiResponse)
	Delete(org cf.Organization) (apiResponse net.ApiResponse)
	FindQuotaByName(name string) (quota cf.Quota, apiResponse net.ApiResponse)
	UpdateQuota(org cf.Organization, quota cf.Quota) (apiResponse net.ApiResponse)
}

type CloudControllerOrganizationRepository struct {
	config  *configuration.Configuration
	gateway net.Gateway
}

func NewCloudControllerOrganizationRepository(config *configuration.Configuration, gateway net.Gateway) (repo CloudControllerOrganizationRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (repo CloudControllerOrganizationRepository) FindAll() (orgs []cf.Organization, apiResponse net.ApiResponse) {
	path := repo.config.Target + "/v2/organizations"
	request, apiResponse := repo.gateway.NewRequest("GET", path, repo.config.AccessToken, nil)
	if apiResponse.IsNotSuccessful() {
		return
	}
	response := new(PaginatedOrganizationResources)

	_, apiResponse = repo.gateway.PerformRequestForJSONResponse(request, response)

	if apiResponse.IsNotSuccessful() {
		return
	}

	for _, r := range response.Resources {
		orgs = append(orgs, cf.Organization{
			Name: r.Entity.Name,
			Guid: r.Metadata.Guid,
		},
		)
	}

	return
}

func (repo CloudControllerOrganizationRepository) FindByName(name string) (org cf.Organization, apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/organizations?q=name%s&inline-relations-depth=1", repo.config.Target, "%3A"+strings.ToLower(name))
	request, apiResponse := repo.gateway.NewRequest("GET", path, repo.config.AccessToken, nil)
	if apiResponse.IsNotSuccessful() {
		return
	}
	orgResources := new(PaginatedOrganizationResources)

	_, apiResponse = repo.gateway.PerformRequestForJSONResponse(request, orgResources)

	if apiResponse.IsNotSuccessful() {
		return
	}

	if len(orgResources.Resources) == 0 {
		apiResponse = net.NewNotFoundApiResponse("%s %s not found", "Org", name)
		return
	}

	r := orgResources.Resources[0]
	spaces := []cf.Space{}

	for _, s := range r.Entity.Spaces {
		spaces = append(spaces, cf.Space{Name: s.Entity.Name, Guid: s.Metadata.Guid})
	}

	domains := []cf.Domain{}

	for _, d := range r.Entity.Domains {
		domains = append(domains, cf.Domain{Name: d.Entity.Name, Guid: d.Metadata.Guid})
	}

	org = cf.Organization{
		Name:    r.Entity.Name,
		Guid:    r.Metadata.Guid,
		Spaces:  spaces,
		Domains: domains,
	}

	return
}

func (repo CloudControllerOrganizationRepository) Create(name string) (apiResponse net.ApiResponse) {
	path := repo.config.Target + "/v2/organizations"
	data := fmt.Sprintf(
		`{"name":"%s"}`, name,
	)
	request, apiResponse := repo.gateway.NewRequest("POST", path, repo.config.AccessToken, strings.NewReader(data))
	if apiResponse.IsNotSuccessful() {
		return
	}

	apiResponse = repo.gateway.PerformRequest(request)
	return
}

func (repo CloudControllerOrganizationRepository) Rename(org cf.Organization, name string) (apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/organizations/%s", repo.config.Target, org.Guid)
	data := fmt.Sprintf(`{"name":"%s"}`, name)
	request, apiResponse := repo.gateway.NewRequest("PUT", path, repo.config.AccessToken, strings.NewReader(data))
	if apiResponse.IsNotSuccessful() {
		return
	}

	apiResponse = repo.gateway.PerformRequest(request)
	return
}

func (repo CloudControllerOrganizationRepository) Delete(org cf.Organization) (apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/organizations/%s?recursive=true", repo.config.Target, org.Guid)
	request, apiResponse := repo.gateway.NewRequest("DELETE", path, repo.config.AccessToken, nil)
	if apiResponse.IsNotSuccessful() {
		return
	}

	apiResponse = repo.gateway.PerformRequest(request)
	return
}

func (repo CloudControllerOrganizationRepository) FindQuotaByName(name string) (quota cf.Quota, apiResponse net.ApiResponse) {
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

func (repo CloudControllerOrganizationRepository) UpdateQuota(org cf.Organization, quota cf.Quota) (apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/organizations/%s", repo.config.Target, org.Guid)
	data := fmt.Sprintf(`{"quota_definition_guid":"%s"}`, quota.Guid)
	request, apiResponse := repo.gateway.NewRequest("PUT", path, repo.config.AccessToken, strings.NewReader(data))
	if apiResponse.IsNotSuccessful() {
		return
	}

	apiResponse = repo.gateway.PerformRequest(request)
	return
}
