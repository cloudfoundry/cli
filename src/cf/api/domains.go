package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"fmt"
	"strings"
)

type PaginatedDomainResources struct {
	Resources []DomainResource
}

type DomainResource struct {
	Resource
	Entity DomainEntity
}

type DomainEntity struct {
	Name                   string
	OwningOrganizationGuid string `json:"owning_organization_guid"`
	Spaces                 []Resource
}

type DomainRepository interface {
	FindAllInCurrentSpace() (domains []cf.Domain, apiResponse net.ApiResponse)
	FindAllByOrg(org cf.Organization) (domains []cf.Domain, apiResponse net.ApiResponse)
	FindByNameInCurrentSpace(name string) (domain cf.Domain, apiResponse net.ApiResponse)
	FindByNameInOrg(name string, owningOrg cf.Organization) (domain cf.Domain, apiResponse net.ApiResponse)
	Create(domainToCreate cf.Domain, owningOrg cf.Organization) (createdDomain cf.Domain, apiResponse net.ApiResponse)
	Share(domainToShare cf.Domain) (apiResponse net.ApiResponse)
	Delete(domain cf.Domain) (apiResponse net.ApiResponse)
	Map(domain cf.Domain, space cf.Space) (apiResponse net.ApiResponse)
	Unmap(domain cf.Domain, space cf.Space) (apiResponse net.ApiResponse)
}

type CloudControllerDomainRepository struct {
	config  *configuration.Configuration
	gateway net.Gateway
}

func NewCloudControllerDomainRepository(config *configuration.Configuration, gateway net.Gateway) (repo CloudControllerDomainRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (repo CloudControllerDomainRepository) FindAllInCurrentSpace() (domains []cf.Domain, apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/spaces/%s/domains", repo.config.Target, repo.config.Space.Guid)
	return repo.findAllWithPath(path)
}

func (repo CloudControllerDomainRepository) FindAllByOrg(org cf.Organization) (domains []cf.Domain, apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/organizations/%s/domains?inline-relations-depth=1", repo.config.Target, org.Guid)
	return repo.findAllWithPath(path)
}

func (repo CloudControllerDomainRepository) findAllWithPath(path string) (domains []cf.Domain, apiResponse net.ApiResponse) {
	request, apiResponse := repo.gateway.NewRequest("GET", path, repo.config.AccessToken, nil)
	if apiResponse.IsNotSuccessful() {
		return
	}

	domainResources := new(PaginatedDomainResources)
	_, apiResponse = repo.gateway.PerformRequestForJSONResponse(request, domainResources)
	if apiResponse.IsNotSuccessful() {
		return
	}

	for _, r := range domainResources.Resources {
		domain := cf.Domain{
			Name: r.Entity.Name,
			Guid: r.Metadata.Guid,
		}
		domain.Shared = r.Entity.OwningOrganizationGuid == ""

		for _, space := range r.Entity.Spaces {
			domain.Spaces = append(domain.Spaces, cf.Space{
				Name: space.Entity.Name,
				Guid: space.Metadata.Guid,
			})
		}
		domains = append(domains, domain)
	}

	return
}

func (repo CloudControllerDomainRepository) FindByNameInCurrentSpace(name string) (domain cf.Domain, apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/spaces/%s/domains?q=name%%3A%s", repo.config.Target, repo.config.Space.Guid, name)
	return repo.findOneWithPath(path, name)
}

func (repo CloudControllerDomainRepository) FindByNameInOrg(name string, org cf.Organization) (domain cf.Domain, apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/organizations/%s/domains?inline-relations-depth=1&q=name%%3A%s", repo.config.Target, org.Guid, name)
	return repo.findOneWithPath(path, name)
}

func (repo CloudControllerDomainRepository) findOneWithPath(path, name string) (domain cf.Domain, apiResponse net.ApiResponse) {
	domains, apiResponse := repo.findAllWithPath(path)
	if apiResponse.IsNotSuccessful() {
		return
	}

	if len(domains) == 0 {
		apiResponse = net.NewNotFoundApiResponse("Domain %s not found", name)
		return
	}

	domain = domains[0]
	return
}

func (repo CloudControllerDomainRepository) Create(domainToCreate cf.Domain, owningOrg cf.Organization) (createdDomain cf.Domain, apiResponse net.ApiResponse) {
	path := repo.config.Target + "/v2/domains"
	data := fmt.Sprintf(
		`{"name":"%s","wildcard":true,"owning_organization_guid":"%s"}`, domainToCreate.Name, owningOrg.Guid,
	)

	request, apiResponse := repo.gateway.NewRequest("POST", path, repo.config.AccessToken, strings.NewReader(data))
	if apiResponse.IsNotSuccessful() {
		return
	}

	resource := new(Resource)
	_, apiResponse = repo.gateway.PerformRequestForJSONResponse(request, resource)
	if apiResponse.IsNotSuccessful() {
		return
	}

	createdDomain.Guid = resource.Metadata.Guid
	createdDomain.Name = resource.Entity.Name
	return
}

func (repo CloudControllerDomainRepository) Share(domainToShare cf.Domain) (apiResponse net.ApiResponse) {
	path := repo.config.Target + "/v2/domains"
	data := fmt.Sprintf(`{"name":"%s","wildcard":true}`, domainToShare.Name)

	request, apiResponse := repo.gateway.NewRequest("POST", path, repo.config.AccessToken, strings.NewReader(data))
	if apiResponse.IsNotSuccessful() {
		return
	}

	apiResponse = repo.gateway.PerformRequest(request)

	return
}

func (repo CloudControllerDomainRepository) Delete(domain cf.Domain) (apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/domains/%s?recursive=true", repo.config.Target, domain.Guid)
	request, apiResponse := repo.gateway.NewRequest("DELETE", path, repo.config.AccessToken, nil)
	if apiResponse.IsNotSuccessful() {
		return
	}

	apiResponse = repo.gateway.PerformRequest(request)
	return
}

func (repo CloudControllerDomainRepository) Map(domain cf.Domain, space cf.Space) (apiResponse net.ApiResponse) {
	return repo.mapOrUnmap("PUT", domain, space)
}

func (repo CloudControllerDomainRepository) Unmap(domain cf.Domain, space cf.Space) (apiResponse net.ApiResponse) {
	return repo.mapOrUnmap("DELETE", domain, space)
}

func (repo CloudControllerDomainRepository) mapOrUnmap(verb string, domain cf.Domain, space cf.Space) (apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/spaces/%s/domains/%s", repo.config.Target, space.Guid, domain.Guid)

	request, apiResponse := repo.gateway.NewRequest(verb, path, repo.config.AccessToken, nil)
	if apiResponse.IsNotSuccessful() {
		return
	}

	apiResponse = repo.gateway.PerformRequest(request)
	return
}
