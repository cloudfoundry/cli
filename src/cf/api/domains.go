package api

import (
	"cf/api/resources"
	"cf/configuration"
	"cf/errors"
	"cf/models"
	"cf/net"
	"fmt"
	"net/url"
	"strings"
)

type DomainRepository interface {
	ListDomainsForOrg(orgGuid string, cb func(models.DomainFields) bool) error
	FindByName(name string) (domain models.DomainFields, apiErr error)
	FindByNameInOrg(name string, owningOrgGuid string) (domain models.DomainFields, apiErr error)
	Create(domainName string, owningOrgGuid string) (createdDomain models.DomainFields, apiErr error)
	CreateSharedDomain(domainName string) (apiErr error)
	Delete(domainGuid string) (apiErr error)
	DeleteSharedDomain(domainGuid string) (apiErr error)
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

func (repo CloudControllerDomainRepository) ListDomainsForOrg(orgGuid string, cb func(models.DomainFields) bool) error {
	apiErr := repo.listDomains(fmt.Sprintf("/v2/organizations/%s/domains", orgGuid), cb)

	// FIXME: needs semantic versioning
	switch apiErr.(type) {
	case *errors.HttpNotFoundError:
		apiErr = repo.listDomains("/v2/domains", cb)
	}

	return apiErr
}

func (repo CloudControllerDomainRepository) listDomains(path string, cb func(models.DomainFields) bool) (apiErr error) {
	return repo.gateway.ListPaginatedResources(
		repo.config.ApiEndpoint(),
		repo.config.AccessToken(),
		path,
		resources.DomainResource{},
		func(resource interface{}) bool {
			return cb(resource.(resources.DomainResource).ToFields())
		})
}

func (repo CloudControllerDomainRepository) isOrgDomain(orgGuid string, domain models.DomainFields) bool {
	return orgGuid == domain.OwningOrganizationGuid || domain.Shared
}

func (repo CloudControllerDomainRepository) FindByName(name string) (domain models.DomainFields, apiErr error) {
	return repo.findOneWithPath(
		fmt.Sprintf("/v2/domains?inline-relations-depth=1&q=%s", url.QueryEscape("name:"+name)),
		name)
}

func (repo CloudControllerDomainRepository) FindByNameInOrg(name string, orgGuid string) (domain models.DomainFields, apiErr error) {
	domain, apiErr = repo.findOneWithPath(
		fmt.Sprintf("/v2/organizations/%s/domains?inline-relations-depth=1&q=%s", orgGuid, url.QueryEscape("name:"+name)),
		name)

	switch apiErr.(type) {
	case *errors.ModelNotFoundError:
		domain, apiErr = repo.FindByName(name)
		if !domain.Shared {
			apiErr = errors.NewModelNotFoundError("Domain", name)
		}
	}

	return
}

func (repo CloudControllerDomainRepository) findOneWithPath(path, name string) (domain models.DomainFields, apiErr error) {
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

func (repo CloudControllerDomainRepository) Create(domainName string, owningOrgGuid string) (createdDomain models.DomainFields, apiErr error) {
	data := fmt.Sprintf(`{"name":"%s","owning_organization_guid":"%s"}`, domainName, owningOrgGuid)
	resource := new(resources.DomainResource)

	path := repo.config.ApiEndpoint() + "/v2/private_domains"
	apiErr = repo.gateway.CreateResourceForResponse(path, repo.config.AccessToken(), strings.NewReader(data), resource)

	switch apiErr.(type) {
	case *errors.HttpNotFoundError:
		path := repo.config.ApiEndpoint() + "/v2/domains"
		data := fmt.Sprintf(`{"name":"%s","owning_organization_guid":"%s", "wildcard": true}`, domainName, owningOrgGuid)
		apiErr = repo.gateway.CreateResourceForResponse(path, repo.config.AccessToken(), strings.NewReader(data), resource)
	}

	if apiErr == nil {
		createdDomain = resource.ToFields()
	}
	return
}

func (repo CloudControllerDomainRepository) CreateSharedDomain(domainName string) (apiErr error) {
	path := repo.config.ApiEndpoint() + "/v2/shared_domains"
	data := strings.NewReader(fmt.Sprintf(`{"name":"%s"}`, domainName))
	apiErr = repo.gateway.CreateResource(path, repo.config.AccessToken(), data)

	switch apiErr.(type) {
	case *errors.HttpNotFoundError:
		path := repo.config.ApiEndpoint() + "/v2/domains"
		data := strings.NewReader(fmt.Sprintf(`{"name":"%s", "wildcard": true}`, domainName))
		apiErr = repo.gateway.CreateResource(path, repo.config.AccessToken(), data)
	}
	return
}

func (repo CloudControllerDomainRepository) Delete(domainGuid string) (apiErr error) {
	path := fmt.Sprintf("%s/v2/private_domains/%s?recursive=true", repo.config.ApiEndpoint(), domainGuid)
	apiErr = repo.gateway.DeleteResource(path, repo.config.AccessToken())

	switch apiErr.(type) {
	case *errors.HttpNotFoundError:
		path := fmt.Sprintf("%s/v2/domains/%s?recursive=true", repo.config.ApiEndpoint(), domainGuid)
		apiErr = repo.gateway.DeleteResource(path, repo.config.AccessToken())
	}
	return
}

func (repo CloudControllerDomainRepository) DeleteSharedDomain(domainGuid string) (apiErr error) {
	path := fmt.Sprintf("%s/v2/shared_domains/%s?recursive=true", repo.config.ApiEndpoint(), domainGuid)
	apiErr = repo.gateway.DeleteResource(path, repo.config.AccessToken())

	switch apiErr.(type) {
	case *errors.HttpNotFoundError:
		path := fmt.Sprintf("%s/v2/domains/%s?recursive=true", repo.config.ApiEndpoint(), domainGuid)
		apiErr = repo.gateway.DeleteResource(path, repo.config.AccessToken())
	}
	return
}
