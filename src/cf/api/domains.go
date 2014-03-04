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
	FindByName(name string) (domain models.DomainFields, apiErr errors.Error)
	FindByNameInOrg(name string, owningOrgGuid string) (domain models.DomainFields, apiErr errors.Error)
	Create(domainName string, owningOrgGuid string) (createdDomain models.DomainFields, apiErr errors.Error)
	CreateSharedDomain(domainName string) (apiErr errors.Error)
	Delete(domainGuid string) (apiErr errors.Error)
	DeleteSharedDomain(domainGuid string) (apiErr errors.Error)
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
	apiErr := repo.listDomains(fmt.Sprintf("/v2/organizations/%s/private_domains", orgGuid), cb)

	// FIXME: needs semantic versioning
	switch apiErr.(type) {
	case errors.HttpNotFoundError:
		apiErr = repo.listDomains("/v2/domains", cb)
	}

	return apiErr
}

func (repo CloudControllerDomainRepository) listDomains(path string, cb func(models.DomainFields) bool) (apiErr errors.Error) {
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

func (repo CloudControllerDomainRepository) FindByName(name string) (domain models.DomainFields, apiErr errors.Error) {
	return repo.findOneWithPath(
		fmt.Sprintf("/v2/domains?inline-relations-depth=1&q=%s", url.QueryEscape("name:"+name)),
		name)
}

func (repo CloudControllerDomainRepository) FindByNameInOrg(name string, orgGuid string) (domain models.DomainFields, apiErr errors.Error) {
	domain, apiErr = repo.findOneWithPath(
		fmt.Sprintf("/v2/organizations/%s/domains?inline-relations-depth=1&q=%s", orgGuid, url.QueryEscape("name:"+name)),
		name)

	switch apiErr.(type) {
	case errors.ModelNotFoundError:
		domain, apiErr = repo.FindByName(name)
		if !domain.Shared {
			apiErr = errors.NewModelNotFoundError("Domain", name)
		}
	}

	return
}

func (repo CloudControllerDomainRepository) findOneWithPath(path, name string) (domain models.DomainFields, apiErr errors.Error) {
	foundDomain := false
	apiErr = repo.listDomains(path, func(result models.DomainFields) bool {
		domain = result
		foundDomain = true
		return false
	})

	if apiErr == nil && !foundDomain {
		apiErr = errors.NewModelNotFoundError("Domain", name)
	}

	return
}

func (repo CloudControllerDomainRepository) Create(domainName string, owningOrgGuid string) (createdDomain models.DomainFields, apiErr errors.Error) {
	data := fmt.Sprintf(`{"name":"%s","owning_organization_guid":"%s"}`, domainName, owningOrgGuid)
	resource := new(DomainResource)

	path := repo.config.ApiEndpoint() + "/v2/private_domains"
	apiErr = repo.gateway.CreateResourceForResponse(path, repo.config.AccessToken(), strings.NewReader(data), resource)

	switch apiErr.(type) {
	case errors.HttpNotFoundError:
		path := repo.config.ApiEndpoint() + "/v2/domains"
		data := fmt.Sprintf(`{"name":"%s","owning_organization_guid":"%s", "wildcard": true}`, domainName, owningOrgGuid)
		apiErr = repo.gateway.CreateResourceForResponse(path, repo.config.AccessToken(), strings.NewReader(data), resource)
	}

	if apiErr == nil {
		createdDomain = resource.ToFields()
	}
	return
}

func (repo CloudControllerDomainRepository) CreateSharedDomain(domainName string) (apiErr errors.Error) {
	path := repo.config.ApiEndpoint() + "/v2/shared_domains"
	data := strings.NewReader(fmt.Sprintf(`{"name":"%s"}`, domainName))
	apiErr = repo.gateway.CreateResource(path, repo.config.AccessToken(), data)

	switch apiErr.(type) {
	case errors.HttpNotFoundError:
		path := repo.config.ApiEndpoint() + "/v2/domains"
		data := strings.NewReader(fmt.Sprintf(`{"name":"%s", "wildcard": true}`, domainName))
		apiErr = repo.gateway.CreateResource(path, repo.config.AccessToken(), data)
	}
	return
}

func (repo CloudControllerDomainRepository) Delete(domainGuid string) (apiErr errors.Error) {
	path := fmt.Sprintf("%s/v2/private_domains/%s?recursive=true", repo.config.ApiEndpoint(), domainGuid)
	apiErr = repo.gateway.DeleteResource(path, repo.config.AccessToken())

	switch apiErr.(type) {
	case errors.HttpNotFoundError:
		path := fmt.Sprintf("%s/v2/domains/%s?recursive=true", repo.config.ApiEndpoint(), domainGuid)
		apiErr = repo.gateway.DeleteResource(path, repo.config.AccessToken())
	}
	return
}

func (repo CloudControllerDomainRepository) DeleteSharedDomain(domainGuid string) (apiErr errors.Error) {
	path := fmt.Sprintf("%s/v2/shared_domains/%s?recursive=true", repo.config.ApiEndpoint(), domainGuid)
	apiErr = repo.gateway.DeleteResource(path, repo.config.AccessToken())

	switch apiErr.(type) {
	case errors.HttpNotFoundError:
		path := fmt.Sprintf("%s/v2/domains/%s?recursive=true", repo.config.ApiEndpoint(), domainGuid)
		apiErr = repo.gateway.DeleteResource(path, repo.config.AccessToken())
	}
	return
}
