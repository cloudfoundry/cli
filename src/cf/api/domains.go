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
	FindAllByOrg(org cf.Organization) (domains []cf.Domain, apiResponse net.ApiResponse)
	FindByNameInCurrentSpace(name string) (domain cf.Domain, apiResponse net.ApiResponse)
	FindByNameInOrg(name string, owningOrg cf.Organization) (domain cf.Domain, apiResponse net.ApiResponse)
	Create(domainToCreate cf.Domain, owningOrg cf.Organization) (createdDomain cf.Domain, apiResponse net.ApiResponse)
	CreateSharedDomain(domainToShare cf.Domain) (apiResponse net.ApiResponse)
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

func (repo CloudControllerDomainRepository) FindAllByOrg(org cf.Organization) (domains []cf.Domain, apiResponse net.ApiResponse) {
	scopedPath := fmt.Sprintf("%s/v2/organizations/%s/domains?inline-relations-depth=1", repo.config.Target, org.Guid)
	domains, apiResponse = repo.findAllWithPath(scopedPath)
	if apiResponse.IsNotSuccessful() {
		return
	}

	sharedPath := fmt.Sprintf("%s/v2/domains?inline-relations-depth=1", repo.config.Target)
	sharedDomains, apiResponse := repo.findAllWithPath(sharedPath)
	if apiResponse.IsNotSuccessful() {
		return
	}

	var domainIsNotIncluded = func(domain cf.Domain) bool {
		for _, d := range domains {
			if d.Guid == domain.Guid {
				return false
			}
		}
		return true
	}

	for _, d := range sharedDomains {
		if domainIsNotIncluded(d) {
			domains = append(domains, d)
		}
	}
	return
}

func (repo CloudControllerDomainRepository) findAllWithPath(path string) (domains []cf.Domain, apiResponse net.ApiResponse) {
	domainResources := new(PaginatedDomainResources)
	apiResponse = repo.gateway.GetResource(path, repo.config.AccessToken, domainResources)
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
	spacePath := fmt.Sprintf("%s/v2/spaces/%s/domains?q=name%%3A%s", repo.config.Target, repo.config.Space.Guid, name)
	sharedPath := fmt.Sprintf("%s/v2/domains?q=name%%3A%s", repo.config.Target, name)
	return repo.findOneWithPaths(spacePath, sharedPath, name)
}

func (repo CloudControllerDomainRepository) FindByNameInOrg(name string, org cf.Organization) (domain cf.Domain, apiResponse net.ApiResponse) {
	orgPath := fmt.Sprintf("%s/v2/organizations/%s/domains?inline-relations-depth=1&q=name%%3A%s", repo.config.Target, org.Guid, name)
	sharedPath := fmt.Sprintf("%s/v2/domains?inline-relations-depth=1&q=name%%3A%s", repo.config.Target, name)
	return repo.findOneWithPaths(orgPath, sharedPath, name)
}

func (repo CloudControllerDomainRepository) findOneWithPaths(scopedPath, sharedPath, name string) (domain cf.Domain, apiResponse net.ApiResponse) {
	domains, apiResponse := repo.findAllWithPath(scopedPath)
	if apiResponse.IsNotSuccessful() {
		return
	}

	if len(domains) == 0 {
		domains, apiResponse = repo.findAllWithPath(sharedPath)
		if apiResponse.IsNotSuccessful() {
			return
		}

		if len(domains) == 0 || !domains[0].Shared {
			apiResponse = net.NewNotFoundApiResponse("Domain %s not found", name)
			return
		}
	}

	domain = domains[0]
	return
}

func (repo CloudControllerDomainRepository) Create(domainToCreate cf.Domain, owningOrg cf.Organization) (createdDomain cf.Domain, apiResponse net.ApiResponse) {
	path := repo.config.Target + "/v2/domains"
	data := fmt.Sprintf(
		`{"name":"%s","wildcard":true,"owning_organization_guid":"%s"}`, domainToCreate.Name, owningOrg.Guid,
	)

	resource := new(Resource)
	apiResponse = repo.gateway.CreateResourceForResponse(path, repo.config.AccessToken, strings.NewReader(data), resource)
	if apiResponse.IsNotSuccessful() {
		return
	}

	createdDomain.Guid = resource.Metadata.Guid
	createdDomain.Name = resource.Entity.Name
	return
}

func (repo CloudControllerDomainRepository) CreateSharedDomain(domain cf.Domain) (apiResponse net.ApiResponse) {
	path := repo.config.Target + "/v2/domains"
	data := fmt.Sprintf(`{"name":"%s","wildcard":true}`, domain.Name)
	return repo.gateway.CreateResource(path, repo.config.AccessToken, strings.NewReader(data))
}

func (repo CloudControllerDomainRepository) Delete(domain cf.Domain) (apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/domains/%s?recursive=true", repo.config.Target, domain.Guid)
	return repo.gateway.DeleteResource(path, repo.config.AccessToken)
}

func (repo CloudControllerDomainRepository) Map(domain cf.Domain, space cf.Space) (apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/spaces/%s/domains/%s", repo.config.Target, space.Guid, domain.Guid)
	return repo.gateway.UpdateResource(path, repo.config.AccessToken, nil)
}

func (repo CloudControllerDomainRepository) Unmap(domain cf.Domain, space cf.Space) (apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/spaces/%s/domains/%s", repo.config.Target, space.Guid, domain.Guid)
	return repo.gateway.DeleteResource(path, repo.config.AccessToken)
}
