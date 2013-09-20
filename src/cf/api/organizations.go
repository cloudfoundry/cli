package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"fmt"
	"strings"
)

type OrganizationRepository interface {
	FindAll() (orgs []cf.Organization, apiErr *net.ApiError)
	FindByName(name string) (org cf.Organization, apiErr *net.ApiError)
	Create(name string) (apiErr *net.ApiError)
	Rename(org cf.Organization, name string) (apiErr *net.ApiError)
	Delete(org cf.Organization) (apiErr *net.ApiError)
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

func (repo CloudControllerOrganizationRepository) FindAll() (orgs []cf.Organization, apiErr *net.ApiError) {
	path := repo.config.Target + "/v2/organizations?inline-relations-depth=1"
	request, apiErr := repo.gateway.NewRequest("GET", path, repo.config.AccessToken, nil)
	if apiErr != nil {
		return
	}
	response := new(OrganizationsApiResponse)

	apiErr = repo.gateway.PerformRequestForJSONResponse(request, response)

	if apiErr != nil {
		return
	}

	for _, r := range response.Resources {
		spaces := []cf.Space{}

		for _, s := range r.Entity.Spaces {
			spaces = append(spaces, cf.Space{Name: s.Entity.Name, Guid: s.Metadata.Guid})
		}

		domains := []cf.Domain{}

		for _, d := range r.Entity.Domains {
			domains = append(domains, cf.Domain{Name: d.Entity.Name, Guid: d.Metadata.Guid})
		}

		orgs = append(orgs, cf.Organization{
			Name:    r.Entity.Name,
			Guid:    r.Metadata.Guid,
			Spaces:  spaces,
			Domains: domains,
		},
		)
	}

	return
}

func (repo CloudControllerOrganizationRepository) FindByName(name string) (org cf.Organization, apiErr *net.ApiError) {
	orgs, apiErr := repo.FindAll()
	lowerName := strings.ToLower(name)

	if apiErr != nil {
		return
	}

	for _, o := range orgs {
		if strings.ToLower(o.Name) == lowerName {
			return o, nil
		}
	}

	apiErr = net.NewApiErrorWithMessage("Organization not found")
	return
}

func (repo CloudControllerOrganizationRepository) Create(name string) (apiErr *net.ApiError) {
	path := repo.config.Target + "/v2/organizations"
	data := fmt.Sprintf(
		`{"name":"%s"}`, name,
	)
	request, apiErr := repo.gateway.NewRequest("POST", path, repo.config.AccessToken, strings.NewReader(data))
	if apiErr != nil {
		return
	}

	apiErr = repo.gateway.PerformRequest(request)
	return
}

func (repo CloudControllerOrganizationRepository) Rename(org cf.Organization, name string) (apiErr *net.ApiError) {
	path := fmt.Sprintf("%s/v2/organizations/%s", repo.config.Target, org.Guid)
	data := fmt.Sprintf(`{"name":"%s"}`, name)
	request, apiErr := repo.gateway.NewRequest("PUT", path, repo.config.AccessToken, strings.NewReader(data))
	if apiErr != nil {
		return
	}

	apiErr = repo.gateway.PerformRequest(request)
	return
}

func (repo CloudControllerOrganizationRepository) Delete(org cf.Organization) (apiErr *net.ApiError) {
	path := fmt.Sprintf("%s/v2/organizations/%s?recursive=true", repo.config.Target, org.Guid)
	request, apiErr := repo.gateway.NewRequest("DELETE", path, repo.config.AccessToken, nil)
	if apiErr != nil {
		return
	}

	apiErr = repo.gateway.PerformRequest(request)
	return
}
