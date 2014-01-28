package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"fmt"
	"net/url"
	"strings"
)

type PaginatedDomainResources struct {
	NextUrl   string `json:"next_url"`
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
	return
}

type DomainEntity struct {
	Name                   string
	OwningOrganizationGuid string `json:"owning_organization_guid"`
}

type ListDomainsCallback func(domains []cf.Domain) (fetchNext bool)

type DomainRepository interface {
	ListDomainsForOrg(orgGuid string, cb ListDomainsCallback) net.ApiResponse
	ListSharedDomains(cb ListDomainsCallback) net.ApiResponse
	FindByName(name string) (domain cf.Domain, apiResponse net.ApiResponse)
	FindByNameInCurrentSpace(name string) (domain cf.Domain, apiResponse net.ApiResponse)
	FindByNameInOrg(name string, owningOrgGuid string) (domain cf.Domain, apiResponse net.ApiResponse)
	Create(domainName string, owningOrgGuid string) (createdDomain cf.DomainFields, apiResponse net.ApiResponse)
	CreateSharedDomain(domainName string) (apiResponse net.ApiResponse)
	Delete(domainGuid string) (apiResponse net.ApiResponse)
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

func (repo CloudControllerDomainRepository) ListSharedDomains(cb ListDomainsCallback) net.ApiResponse {
	return repo.listDomains("/v2/shared_domains?inline-relations-depth=1", cb)
}

func (repo CloudControllerDomainRepository) ListDomainsForOrg(orgGuid string, cb ListDomainsCallback) net.ApiResponse {
	return repo.listDomains(fmt.Sprintf("/v2/organizations/%s/private_domains", orgGuid), cb)
}

func (repo CloudControllerDomainRepository) listDomains(path string, cb ListDomainsCallback) (apiResponse net.ApiResponse) {
	fetchNext := true
	for fetchNext {
		var (
			domains     []cf.Domain
			shouldFetch bool
		)

		domains, path, apiResponse = repo.findNextWithPath(path)
		if apiResponse.IsNotSuccessful() {
			return
		}

		if len(domains) > 0 {
			shouldFetch = cb(domains)
		}

		fetchNext = shouldFetch && path != ""
	}
	return
}

func (repo CloudControllerDomainRepository) isOrgDomain(orgGuid string, domain cf.DomainFields) bool {
	return orgGuid == domain.OwningOrganizationGuid || domain.Shared
}

func (repo CloudControllerDomainRepository) findNextWithPath(path string) (domains []cf.Domain, nextUrl string, apiResponse net.ApiResponse) {
	domainResources := new(PaginatedDomainResources)

	apiResponse = repo.gateway.GetResource(repo.config.Target+path, repo.config.AccessToken, domainResources)
	if apiResponse.IsNotSuccessful() {
		return
	}

	nextUrl = domainResources.NextUrl
	for _, r := range domainResources.Resources {
		domains = append(domains, r.ToModel())
	}

	return
}

func (repo CloudControllerDomainRepository) FindByName(name string) (domain cf.Domain, apiResponse net.ApiResponse) {
	path := fmt.Sprintf("/v2/domains?inline-relations-depth=1&q=%s", url.QueryEscape("name:"+name))
	domains, _, apiResponse := repo.findNextWithPath(path)
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
	spacePath := fmt.Sprintf("/v2/spaces/%s/domains?inline-relations-depth=1&q=%s", repo.config.SpaceFields.Guid, url.QueryEscape("name:"+name))
	return repo.findOneWithPaths(spacePath, name)
}

func (repo CloudControllerDomainRepository) FindByNameInOrg(name string, orgGuid string) (domain cf.Domain, apiResponse net.ApiResponse) {
	orgPath := fmt.Sprintf("/v2/organizations/%s/domains?inline-relations-depth=1&q=%s", orgGuid, url.QueryEscape("name:"+name))
	return repo.findOneWithPaths(orgPath, name)
}

func (repo CloudControllerDomainRepository) findOneWithPaths(scopedPath, name string) (domain cf.Domain, apiResponse net.ApiResponse) {
	domains, _, apiResponse := repo.findNextWithPath(scopedPath)
	if apiResponse.IsNotSuccessful() {
		return
	}

	if len(domains) == 0 {
		sharedPath := fmt.Sprintf("/v2/domains?inline-relations-depth=1&q=%s", url.QueryEscape("name:"+name))
		domains, _, apiResponse = repo.findNextWithPath(sharedPath)
		if apiResponse.IsNotSuccessful() {
			return
		}

		if len(domains) == 0 || !domains[0].Shared {
			apiResponse = net.NewNotFoundApiResponse("Domain '%s' not found", name)
			return
		}
	}

	domain = domains[0]
	return
}

func (repo CloudControllerDomainRepository) Create(domainName string, owningOrgGuid string) (createdDomain cf.DomainFields, apiResponse net.ApiResponse) {
	data := fmt.Sprintf(`{"name":"%s","owning_organization_guid":"%s"}`, domainName, owningOrgGuid)
	resource := new(DomainResource)

	path := repo.config.Target + "/v2/private_domains"
	apiResponse = repo.gateway.CreateResourceForResponse(path, repo.config.AccessToken, strings.NewReader(data), resource)

	if apiResponse.IsNotFound() {
		path := repo.config.Target + "/v2/domains"
		apiResponse = repo.gateway.CreateResourceForResponse(path, repo.config.AccessToken, strings.NewReader(data), resource)
	}

	if apiResponse.IsSuccessful() {
		createdDomain = resource.ToFields()
	}
	return
}

func (repo CloudControllerDomainRepository) CreateSharedDomain(domainName string) (apiResponse net.ApiResponse) {
	path := repo.config.Target + "/v2/shared_domains"
	data := strings.NewReader(fmt.Sprintf(`{"name":"%s"}`, domainName))
	apiResponse = repo.gateway.CreateResource(path, repo.config.AccessToken, data)

	if apiResponse.IsNotFound() {
		path := repo.config.Target + "/v2/domains"
		apiResponse = repo.gateway.CreateResource(path, repo.config.AccessToken, data)
	}
	return
}

func (repo CloudControllerDomainRepository) Delete(domainGuid string) (apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/private_domains/%s?recursive=true", repo.config.Target, domainGuid)
	apiResponse = repo.gateway.DeleteResource(path, repo.config.AccessToken)

	if apiResponse.IsNotFound() {
		path := fmt.Sprintf("%s/v2/domains/%s?recursive=true", repo.config.Target, domainGuid)
		apiResponse = repo.gateway.DeleteResource(path, repo.config.AccessToken)
	}
	return
}
