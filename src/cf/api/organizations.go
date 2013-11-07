package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"cf/paginator"
	"fmt"
	"strings"
)

type OrganizationResources struct {
	Resources []OrganizationResource
	NextUrl   string
}

type OrganizationResource struct {
	Resource
	Entity OrganizationEntity
}

type OrganizationEntity struct {
	Name    string
	Spaces  []Resource
	Domains []Resource
}

type OrganizationRepository interface {
	FindAll() (orgs []cf.Organization, apiResponse net.ApiResponse)
	Paginator() (paginator paginator.Paginator)
	FindByName(name string) (org cf.Organization, apiResponse net.ApiResponse)
	Create(name string) (apiResponse net.ApiResponse)
	Rename(org cf.Organization, name string) (apiResponse net.ApiResponse)
	Delete(org cf.Organization) (apiResponse net.ApiResponse)
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
	return repo.findAllWithPath(path)
}

func (repo CloudControllerOrganizationRepository) Paginator() paginator.Paginator {
	return paginator.NewOrganizationPaginator(repo.config, repo.gateway)
}

func (repo CloudControllerOrganizationRepository) FindByName(name string) (org cf.Organization, apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/organizations?q=name%s&inline-relations-depth=1", repo.config.Target, "%3A"+strings.ToLower(name))

	orgs, apiResponse := repo.findAllWithPath(path)
	if apiResponse.IsNotSuccessful() {
		return
	}

	if len(orgs) == 0 {
		apiResponse = net.NewNotFoundApiResponse("Org %s not found", name)
		return
	}

	org = orgs[0]
	return
}

func (repo CloudControllerOrganizationRepository) findAllWithPath(path string) (orgs []cf.Organization, apiResponse net.ApiResponse) {
	orgResources := new(OrganizationResources)
	apiResponse = repo.gateway.GetResource(path, repo.config.AccessToken, orgResources)
	if apiResponse.IsNotSuccessful() {
		return
	}

	for _, r := range orgResources.Resources {
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
		})
	}
	return
}

func (repo CloudControllerOrganizationRepository) Create(name string) (apiResponse net.ApiResponse) {
	url := repo.config.Target + "/v2/organizations"
	data := fmt.Sprintf(`{"name":"%s"}`, name)
	return repo.gateway.CreateResource(url, repo.config.AccessToken, strings.NewReader(data))
}

func (repo CloudControllerOrganizationRepository) Rename(org cf.Organization, name string) (apiResponse net.ApiResponse) {
	url := fmt.Sprintf("%s/v2/organizations/%s", repo.config.Target, org.Guid)
	data := fmt.Sprintf(`{"name":"%s"}`, name)
	return repo.gateway.UpdateResource(url, repo.config.AccessToken, strings.NewReader(data))
}

func (repo CloudControllerOrganizationRepository) Delete(org cf.Organization) (apiResponse net.ApiResponse) {
	url := fmt.Sprintf("%s/v2/organizations/%s?recursive=true", repo.config.Target, org.Guid)
	return repo.gateway.DeleteResource(url, repo.config.AccessToken)
}
