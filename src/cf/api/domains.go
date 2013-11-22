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

func (resource DomainResource) ToFields() (fields cf.DomainFields) {
	fields.Name = resource.Entity.Name
	fields.Guid = resource.Metadata.Guid
	fields.OwningOrganizationGuid = resource.Entity.OwningOrganizationGuid
	fields.Shared = fields.OwningOrganizationGuid == ""
	return
}

func (resource DomainResource) ToModel() (domain cf.Domain) {
	domain.DomainFields = resource.ToFields()

	for _, spaceResource := range resource.Entity.Spaces {
		domain.Spaces = append(domain.Spaces, spaceResource.ToFields())
	}

	return
}

type DomainEntity struct {
	Name                   string
	OwningOrganizationGuid string `json:"owning_organization_guid"`
	Spaces                 []SpaceResource
}

type DomainRepository interface {
	FindDefaultAppDomain() (domain cf.Domain, apiResponse net.ApiResponse)
	FindAllByOrg(orgGuid string) (domains []cf.Domain, apiResponse net.ApiResponse)
	FindByName(name string) (domain cf.Domain, apiResponse net.ApiResponse)
	FindByNameInCurrentSpace(name string) (domain cf.Domain, apiResponse net.ApiResponse)
	FindByNameInOrg(name string, owningOrgGuid string) (domain cf.Domain, apiResponse net.ApiResponse)
	Create(domainName string, owningOrgGuid string) (createdDomain cf.DomainFields, apiResponse net.ApiResponse)
	CreateSharedDomain(domainName string) (apiResponse net.ApiResponse)
	Delete(domainGuid string) (apiResponse net.ApiResponse)
	Map(domainGuid string, spaceGuid string) (apiResponse net.ApiResponse)
	Unmap(domainGuid string, spaceGuid string) (apiResponse net.ApiResponse)
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

func (repo CloudControllerDomainRepository) FindDefaultAppDomain() (domain cf.Domain, apiResponse net.ApiResponse) {
	sharedPath := fmt.Sprintf("%s/v2/domains?inline-relations-depth=1", repo.config.Target)
	sharedDomains, apiResponse := repo.findAllWithPath(sharedPath)
	if apiResponse.IsNotSuccessful() {
		return
	}

	if len(sharedDomains) > 0 {
		domain = sharedDomains[0]
	} else {
		apiResponse = net.NewNotFoundApiResponse("No default domain exists")
	}

	return
}

func (repo CloudControllerDomainRepository) FindAllByOrg(orgGuid string) (domains []cf.Domain, apiResponse net.ApiResponse) {
	sharedPath := fmt.Sprintf("%s/v2/domains?inline-relations-depth=1", repo.config.Target)
	allDomains, apiResponse := repo.findAllWithPath(sharedPath)
	if apiResponse.IsNotSuccessful() {
		return
	}

	for _, d := range allDomains {
		if repo.isOrgDomain(orgGuid, d.DomainFields) {
			domains = append(domains, d)
		}
	}
	return
}

func (repo CloudControllerDomainRepository) isOrgDomain(orgGuid string, domain cf.DomainFields) bool {
	return orgGuid == domain.OwningOrganizationGuid || domain.Shared
}

func (repo CloudControllerDomainRepository) findAllWithPath(path string) (domains []cf.Domain, apiResponse net.ApiResponse) {
	domainResources := new(PaginatedDomainResources)
	apiResponse = repo.gateway.GetResource(path, repo.config.AccessToken, domainResources)
	if apiResponse.IsNotSuccessful() {
		return
	}

	for _, r := range domainResources.Resources {
		domains = append(domains, r.ToModel())
	}

	return
}

func (repo CloudControllerDomainRepository) FindByName(name string) (domain cf.Domain, apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/domains?inline-relations-depth=1&q=name%%3A%s", repo.config.Target, name)
	domains, apiResponse := repo.findAllWithPath(path)
	if apiResponse.IsNotSuccessful() {
		return
	}

	if len(domains) > 0 {
		domain = domains[0]
	} else {
		apiResponse = net.NewNotFoundApiResponse("Domain %s not found", name)
	}
	return
}

func (repo CloudControllerDomainRepository) FindByNameInCurrentSpace(name string) (domain cf.Domain, apiResponse net.ApiResponse) {
	spacePath := fmt.Sprintf("%s/v2/spaces/%s/domains?inline-relations-depth=1&q=name%%3A%s", repo.config.Target, repo.config.SpaceFields.Guid, name)
	return repo.findOneWithPaths(spacePath, name)
}

func (repo CloudControllerDomainRepository) FindByNameInOrg(name string, orgGuid string) (domain cf.Domain, apiResponse net.ApiResponse) {
	orgPath := fmt.Sprintf("%s/v2/organizations/%s/domains?inline-relations-depth=1&q=name%%3A%s", repo.config.Target, orgGuid, name)
	return repo.findOneWithPaths(orgPath, name)
}

func (repo CloudControllerDomainRepository) findOneWithPaths(scopedPath, name string) (domain cf.Domain, apiResponse net.ApiResponse) {
	domains, apiResponse := repo.findAllWithPath(scopedPath)
	if apiResponse.IsNotSuccessful() {
		return
	}

	if len(domains) == 0 {
		sharedPath := fmt.Sprintf("%s/v2/domains?inline-relations-depth=1&q=name%%3A%s", repo.config.Target, name)
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

func (repo CloudControllerDomainRepository) Create(domainName string, owningOrgGuid string) (createdDomain cf.DomainFields, apiResponse net.ApiResponse) {
	path := repo.config.Target + "/v2/domains"
	data := fmt.Sprintf(
		`{"name":"%s","wildcard":true,"owning_organization_guid":"%s"}`, domainName, owningOrgGuid,
	)

	resource := new(DomainResource)
	apiResponse = repo.gateway.CreateResourceForResponse(path, repo.config.AccessToken, strings.NewReader(data), resource)
	if apiResponse.IsNotSuccessful() {
		return
	}

	createdDomain = resource.ToFields()
	return
}

func (repo CloudControllerDomainRepository) CreateSharedDomain(domainName string) (apiResponse net.ApiResponse) {
	path := repo.config.Target + "/v2/domains"
	data := fmt.Sprintf(`{"name":"%s","wildcard":true}`, domainName)
	return repo.gateway.CreateResource(path, repo.config.AccessToken, strings.NewReader(data))
}

func (repo CloudControllerDomainRepository) Delete(domainGuid string) (apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/domains/%s?recursive=true", repo.config.Target, domainGuid)
	return repo.gateway.DeleteResource(path, repo.config.AccessToken)
}

func (repo CloudControllerDomainRepository) Map(domainGuid string, spaceGuid string) (apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/spaces/%s/domains/%s", repo.config.Target, spaceGuid, domainGuid)
	return repo.gateway.UpdateResource(path, repo.config.AccessToken, nil)
}

func (repo CloudControllerDomainRepository) Unmap(domainGuid string, spaceGuid string) (apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/spaces/%s/domains/%s", repo.config.Target, spaceGuid, domainGuid)
	return repo.gateway.DeleteResource(path, repo.config.AccessToken)
}
