package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"fmt"
	"strings"
)

type OrganizationRepository interface {
	FindAll() (orgs []cf.Organization, apiStatus net.ApiStatus)
	FindByName(name string) (org cf.Organization, apiStatus net.ApiStatus)
	Create(name string) (apiStatus net.ApiStatus)
	Rename(org cf.Organization, name string) (apiStatus net.ApiStatus)
	Delete(org cf.Organization) (apiStatus net.ApiStatus)
	FindQuotaByName(name string) (quota cf.Quota, apiStatus net.ApiStatus)
	UpdateQuota(org cf.Organization, quota cf.Quota) (apiStatus net.ApiStatus)
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

func (repo CloudControllerOrganizationRepository) FindAll() (orgs []cf.Organization, apiStatus net.ApiStatus) {
	path := repo.config.Target + "/v2/organizations"
	request, apiStatus := repo.gateway.NewRequest("GET", path, repo.config.AccessToken, nil)
	if apiStatus.IsError() {
		return
	}
	response := new(OrganizationsApiResponse)

	_, apiStatus = repo.gateway.PerformRequestForJSONResponse(request, response)

	if apiStatus.IsError() {
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

func (repo CloudControllerOrganizationRepository) FindByName(name string) (org cf.Organization, apiStatus net.ApiStatus) {
	path := fmt.Sprintf("%s/v2/organizations?q=name%s&inline-relations-depth=1", repo.config.Target, "%3A"+strings.ToLower(name))
	request, apiStatus := repo.gateway.NewRequest("GET", path, repo.config.AccessToken, nil)
	if apiStatus.IsError() {
		return
	}
	response := new(OrganizationsApiResponse)

	_, apiStatus = repo.gateway.PerformRequestForJSONResponse(request, response)

	if apiStatus.IsError() {
		return
	}

	if len(response.Resources) == 0 {
		apiStatus = net.NewNotFoundApiStatus()
		return
	}

	r := response.Resources[0]
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

func (repo CloudControllerOrganizationRepository) Create(name string) (apiStatus net.ApiStatus) {
	path := repo.config.Target + "/v2/organizations"
	data := fmt.Sprintf(
		`{"name":"%s"}`, name,
	)
	request, apiStatus := repo.gateway.NewRequest("POST", path, repo.config.AccessToken, strings.NewReader(data))
	if apiStatus.IsError() {
		return
	}

	apiStatus = repo.gateway.PerformRequest(request)
	return
}

func (repo CloudControllerOrganizationRepository) Rename(org cf.Organization, name string) (apiStatus net.ApiStatus) {
	path := fmt.Sprintf("%s/v2/organizations/%s", repo.config.Target, org.Guid)
	data := fmt.Sprintf(`{"name":"%s"}`, name)
	request, apiStatus := repo.gateway.NewRequest("PUT", path, repo.config.AccessToken, strings.NewReader(data))
	if apiStatus.IsError() {
		return
	}

	apiStatus = repo.gateway.PerformRequest(request)
	return
}

func (repo CloudControllerOrganizationRepository) Delete(org cf.Organization) (apiStatus net.ApiStatus) {
	path := fmt.Sprintf("%s/v2/organizations/%s?recursive=true", repo.config.Target, org.Guid)
	request, apiStatus := repo.gateway.NewRequest("DELETE", path, repo.config.AccessToken, nil)
	if apiStatus.IsError() {
		return
	}

	apiStatus = repo.gateway.PerformRequest(request)
	return
}


func (repo CloudControllerOrganizationRepository) FindQuotaByName(name string) (quota cf.Quota, apiStatus net.ApiStatus) {
	path := fmt.Sprintf("%s/v2/quota_definitions?q=name%%3A%s", repo.config.Target, name)

	request, apiStatus := repo.gateway.NewRequest("GET", path, repo.config.AccessToken, nil)
	if apiStatus.IsError() {
		return
	}

	response := new(ApiResponse)

	_, apiStatus = repo.gateway.PerformRequestForJSONResponse(request, response)
	if apiStatus.IsError() {
		return
	}

	if len(response.Resources) == 0 {
		apiStatus = net.NewNotFoundApiStatus()
		return
	}

	res := response.Resources[0]
	quota.Guid = res.Metadata.Guid
	quota.Name = res.Entity.Name

	return
}

func (repo CloudControllerOrganizationRepository) UpdateQuota(org cf.Organization, quota cf.Quota) (apiStatus net.ApiStatus) {
	path := fmt.Sprintf("%s/v2/organizations/%s", repo.config.Target, org.Guid)
	data := fmt.Sprintf(`{"quota_definition_guid":"%s"}`, quota.Guid)
	request, apiStatus := repo.gateway.NewRequest("PUT", path, repo.config.AccessToken, strings.NewReader(data))
	if apiStatus.IsError() {
		return
	}

	apiStatus = repo.gateway.PerformRequest(request)
	return
}
