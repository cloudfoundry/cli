package api

import (
	"cf/api/resources"
	"cf/api/strategy"
	"cf/configuration"
	"cf/errors"
	"cf/models"
	"cf/net"
	"encoding/json"
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
	strategy, err := strategy.NewEndpointStrategy(repo.config.ApiVersion())
	if err != nil {
		return err
	}

	return repo.listDomains(strategy.DomainsURL(orgGuid), cb)
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

func (repo CloudControllerDomainRepository) Create(domainName string, owningOrgGuid string) (createdDomain models.DomainFields, err error) {
	strategy, err := strategy.NewEndpointStrategy(repo.config.ApiVersion())
	if err != nil {
		return
	}

	data, err := json.Marshal(resources.DomainEntity{
		Name: domainName,
		OwningOrganizationGuid: owningOrgGuid,
		Wildcard:               true,
	})

	if err != nil {
		return
	}

	resource := new(resources.DomainResource)
	err = repo.gateway.CreateResourceForResponse(
		repo.config.ApiEndpoint()+strategy.PrivateDomainsURL(),
		repo.config.AccessToken(),
		strings.NewReader(string(data)),
		resource)
	if err != nil {
		return
	}

	createdDomain = resource.ToFields()
	return
}

func (repo CloudControllerDomainRepository) CreateSharedDomain(domainName string) (apiErr error) {
	strategy, err := strategy.NewEndpointStrategy(repo.config.ApiVersion())
	if err != nil {
		return
	}

	data, err := json.Marshal(resources.DomainEntity{
		Name:     domainName,
		Wildcard: true,
	})
	if err != nil {
		return
	}

	apiErr = repo.gateway.CreateResource(
		repo.config.ApiEndpoint()+strategy.SharedDomainsURL(),
		repo.config.AccessToken(),
		strings.NewReader(string(data)))

	return
}

func (repo CloudControllerDomainRepository) Delete(domainGuid string) error {
	strategy, err := strategy.NewEndpointStrategy(repo.config.ApiVersion())
	if err != nil {
		return err
	}

	return repo.gateway.DeleteResource(
		repo.config.ApiEndpoint()+strategy.DeleteDomainURL(domainGuid),
		repo.config.AccessToken())
}

func (repo CloudControllerDomainRepository) DeleteSharedDomain(domainGuid string) error {
	strategy, err := strategy.NewEndpointStrategy(repo.config.ApiVersion())
	if err != nil {
		return err
	}

	return repo.gateway.DeleteResource(
		repo.config.ApiEndpoint()+strategy.DeleteSharedDomainURL(domainGuid),
		repo.config.AccessToken())
}
