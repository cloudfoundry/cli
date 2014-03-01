package api

import (
	"cf/configuration"
	"cf/errors"
	"cf/models"
	"cf/net"
	"fmt"
	"net/url"
	"strings"
)

type DomainResource struct {
	Resource
	Entity DomainEntity
}

func (resource DomainResource) ToFields() models.DomainFields {
	owningOrganizationGuid := resource.Entity.OwningOrganizationGuid
	return models.DomainFields{
		Name: resource.Entity.Name,
		Guid: resource.Metadata.Guid,
		OwningOrganizationGuid: owningOrganizationGuid,
		Shared:                 owningOrganizationGuid == "",
	}
}

type DomainEntity struct {
	Name                   string
	OwningOrganizationGuid string `json:"owning_organization_guid"`
}

type DomainRepository interface {
	ListDomainsForOrg(orgGuid string, cb func(models.DomainFields) bool) errors.Error
	ListSharedDomains(cb func(models.DomainFields) bool) errors.Error
	FindByName(name string) (domain models.DomainFields, apiResponse errors.Error)
	FindByNameInOrg(name string, owningOrgGuid string) (domain models.DomainFields, apiResponse errors.Error)
	Create(domainName string, owningOrgGuid string) (createdDomain models.DomainFields, apiResponse errors.Error)
	CreateSharedDomain(domainName string) (apiResponse errors.Error)
	Delete(domainGuid string) (apiResponse errors.Error)
	DeleteSharedDomain(domainGuid string) (apiResponse errors.Error)
	ListDomains(cb func(models.DomainFields) bool) errors.Error
}

type CloudControllerDomainRepository struct {
	config  configuration.Reader
	gateway net.Gateway
}

func NewCloudControllerDomainRepository(config configuration.Reader, gateway net.Gateway) (repo CloudControllerDomainRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (repo CloudControllerDomainRepository) ListSharedDomains(cb func(models.DomainFields) bool) errors.Error {
	return repo.listDomains("/v2/shared_domains", cb)
}

func (repo CloudControllerDomainRepository) ListDomains(cb func(models.DomainFields) bool) errors.Error {
	return repo.listDomains("/v2/domains", cb)
}

func (repo CloudControllerDomainRepository) ListDomainsForOrg(orgGuid string, cb func(models.DomainFields) bool) errors.Error {
	apiResponse := repo.listDomains(fmt.Sprintf("/v2/organizations/%s/private_domains", orgGuid), cb)
	if apiResponse != nil && apiResponse.IsNotFound() { // FIXME: needs semantic versioning
		apiResponse = repo.listDomains("/v2/domains", cb)
	}

	return apiResponse
}

func (repo CloudControllerDomainRepository) listDomains(path string, cb func(models.DomainFields) bool) (apiResponse errors.Error) {
	return repo.gateway.ListPaginatedResources(
		repo.config.ApiEndpoint(),
		repo.config.AccessToken(),
		path,
		DomainResource{},
		func(resource interface{}) bool {
			return cb(resource.(DomainResource).ToFields())
		})
}

func (repo CloudControllerDomainRepository) isOrgDomain(orgGuid string, domain models.DomainFields) bool {
	return orgGuid == domain.OwningOrganizationGuid || domain.Shared
}

func (repo CloudControllerDomainRepository) FindByName(name string) (domain models.DomainFields, apiResponse errors.Error) {
	return repo.findOneWithPath(
		fmt.Sprintf("/v2/domains?inline-relations-depth=1&q=%s", url.QueryEscape("name:"+name)),
		name)
}

func (repo CloudControllerDomainRepository) FindByNameInOrg(name string, orgGuid string) (domain models.DomainFields, apiResponse errors.Error) {
	domain, apiResponse = repo.findOneWithPath(
		fmt.Sprintf("/v2/organizations/%s/domains?inline-relations-depth=1&q=%s", orgGuid, url.QueryEscape("name:"+name)),
		name)

	if apiResponse != nil && apiResponse.IsNotFound() {
		domain, apiResponse = repo.FindByName(name)
		if !domain.Shared {
			apiResponse = errors.NewNotFoundError("Domain %s not found", name)
		}
	}

	return
}

func (repo CloudControllerDomainRepository) findOneWithPath(path, name string) (domain models.DomainFields, apiResponse errors.Error) {
	foundDomain := false
	apiResponse = repo.listDomains(path, func(result models.DomainFields) bool {
		domain = result
		foundDomain = true
		return false
	})

	if apiResponse == nil && !foundDomain {
		apiResponse = errors.NewNotFoundError("Domain %s not found", name)
	}

	return
}

func (repo CloudControllerDomainRepository) Create(domainName string, owningOrgGuid string) (createdDomain models.DomainFields, apiResponse errors.Error) {
	data := fmt.Sprintf(`{"name":"%s","owning_organization_guid":"%s"}`, domainName, owningOrgGuid)
	resource := new(DomainResource)

	path := repo.config.ApiEndpoint() + "/v2/private_domains"
	apiResponse = repo.gateway.CreateResourceForResponse(path, repo.config.AccessToken(), strings.NewReader(data), resource)

	if apiResponse != nil && apiResponse.IsNotFound() {
		path := repo.config.ApiEndpoint() + "/v2/domains"
		data := fmt.Sprintf(`{"name":"%s","owning_organization_guid":"%s", "wildcard": true}`, domainName, owningOrgGuid)
		apiResponse = repo.gateway.CreateResourceForResponse(path, repo.config.AccessToken(), strings.NewReader(data), resource)
	}

	if apiResponse == nil {
		createdDomain = resource.ToFields()
	}
	return
}

func (repo CloudControllerDomainRepository) CreateSharedDomain(domainName string) (apiResponse errors.Error) {
	path := repo.config.ApiEndpoint() + "/v2/shared_domains"
	data := strings.NewReader(fmt.Sprintf(`{"name":"%s"}`, domainName))
	apiResponse = repo.gateway.CreateResource(path, repo.config.AccessToken(), data)

	if apiResponse != nil && apiResponse.IsNotFound() {
		path := repo.config.ApiEndpoint() + "/v2/domains"
		data := strings.NewReader(fmt.Sprintf(`{"name":"%s", "wildcard": true}`, domainName))
		apiResponse = repo.gateway.CreateResource(path, repo.config.AccessToken(), data)
	}
	return
}

func (repo CloudControllerDomainRepository) Delete(domainGuid string) (apiResponse errors.Error) {
	path := fmt.Sprintf("%s/v2/private_domains/%s?recursive=true", repo.config.ApiEndpoint(), domainGuid)
	apiResponse = repo.gateway.DeleteResource(path, repo.config.AccessToken())

	if apiResponse != nil && apiResponse.IsNotFound() {
		path := fmt.Sprintf("%s/v2/domains/%s?recursive=true", repo.config.ApiEndpoint(), domainGuid)
		apiResponse = repo.gateway.DeleteResource(path, repo.config.AccessToken())
	}
	return
}

func (repo CloudControllerDomainRepository) DeleteSharedDomain(domainGuid string) (apiResponse errors.Error) {
	path := fmt.Sprintf("%s/v2/shared_domains/%s?recursive=true", repo.config.ApiEndpoint(), domainGuid)
	apiResponse = repo.gateway.DeleteResource(path, repo.config.AccessToken())

	if apiResponse != nil && apiResponse.IsNotFound() {
		path := fmt.Sprintf("%s/v2/domains/%s?recursive=true", repo.config.ApiEndpoint(), domainGuid)
		apiResponse = repo.gateway.DeleteResource(path, repo.config.AccessToken())
	}
	return
}
