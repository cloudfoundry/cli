package api

import (
	"encoding/json"
	"strings"

	. "github.com/cloudfoundry/cli/cf/i18n"

	"github.com/cloudfoundry/cli/cf/api/resources"
	"github.com/cloudfoundry/cli/cf/api/strategy"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"
)

type DomainRepository interface {
	ListDomainsForOrg(orgGuid string, cb func(models.DomainFields) bool) error
	FindSharedByName(name string) (domain models.DomainFields, apiErr error)
	FindPrivateByName(name string) (domain models.DomainFields, apiErr error)
	FindByNameInOrg(name string, owningOrgGuid string) (domain models.DomainFields, apiErr error)
	Create(domainName string, owningOrgGuid string) (createdDomain models.DomainFields, apiErr error)
	CreateSharedDomain(domainName string, routerGroupGuid string) (apiErr error)
	Delete(domainGuid string) (apiErr error)
	DeleteSharedDomain(domainGuid string) (apiErr error)
	FirstOrDefault(orgGuid string, name *string) (domain models.DomainFields, error error)
}

type CloudControllerDomainRepository struct {
	config   core_config.Reader
	gateway  net.Gateway
	strategy strategy.EndpointStrategy
}

func NewCloudControllerDomainRepository(config core_config.Reader, gateway net.Gateway, strategy strategy.EndpointStrategy) CloudControllerDomainRepository {
	return CloudControllerDomainRepository{
		config:   config,
		gateway:  gateway,
		strategy: strategy,
	}
}

func (repo CloudControllerDomainRepository) ListDomainsForOrg(orgGuid string, cb func(models.DomainFields) bool) error {
	err := repo.listDomains(repo.strategy.PrivateDomainsByOrgURL(orgGuid), cb)
	if err != nil {
		return err
	}
	err = repo.listDomains(repo.strategy.SharedDomainsURL(), cb)
	return err
}

func (repo CloudControllerDomainRepository) listDomains(path string, cb func(models.DomainFields) bool) (apiErr error) {
	return repo.gateway.ListPaginatedResources(
		repo.config.ApiEndpoint(),
		path,
		resources.DomainResource{},
		func(resource interface{}) bool {
			return cb(resource.(resources.DomainResource).ToFields())
		})
}

func (repo CloudControllerDomainRepository) isOrgDomain(orgGuid string, domain models.DomainFields) bool {
	return orgGuid == domain.OwningOrganizationGuid || domain.Shared
}

func (repo CloudControllerDomainRepository) FindSharedByName(name string) (domain models.DomainFields, apiErr error) {
	return repo.findOneWithPath(repo.strategy.SharedDomainURL(name), name)
}

func (repo CloudControllerDomainRepository) FindPrivateByName(name string) (domain models.DomainFields, apiErr error) {
	return repo.findOneWithPath(repo.strategy.PrivateDomainURL(name), name)
}

func (repo CloudControllerDomainRepository) FindByNameInOrg(name string, orgGuid string) (domain models.DomainFields, apiErr error) {
	domain, apiErr = repo.findOneWithPath(repo.strategy.OrgDomainURL(orgGuid, name), name)

	switch apiErr.(type) {
	case *errors.ModelNotFoundError:
		domain, apiErr = repo.FindSharedByName(name)
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
	data, err := json.Marshal(resources.DomainEntity{
		Name: domainName,
		OwningOrganizationGuid: owningOrgGuid,
		Wildcard:               true,
	})

	if err != nil {
		return
	}

	resource := new(resources.DomainResource)
	err = repo.gateway.CreateResource(
		repo.config.ApiEndpoint(),
		repo.strategy.PrivateDomainsURL(),
		strings.NewReader(string(data)),
		resource)

	if err != nil {
		return
	}

	createdDomain = resource.ToFields()
	return
}

func (repo CloudControllerDomainRepository) CreateSharedDomain(domainName string, routerGroupGuid string) (apiErr error) {
	data, err := json.Marshal(resources.DomainEntity{
		Name:            domainName,
		RouterGroupGuid: routerGroupGuid,
		Wildcard:        true,
	})

	if err != nil {
		return
	}

	apiErr = repo.gateway.CreateResource(
		repo.config.ApiEndpoint(),
		repo.strategy.SharedDomainsURL(),
		strings.NewReader(string(data)))

	return
}

func (repo CloudControllerDomainRepository) Delete(domainGuid string) error {
	return repo.gateway.DeleteResource(
		repo.config.ApiEndpoint(),
		repo.strategy.DeleteDomainURL(domainGuid))
}

func (repo CloudControllerDomainRepository) DeleteSharedDomain(domainGuid string) error {
	return repo.gateway.DeleteResource(
		repo.config.ApiEndpoint(),
		repo.strategy.DeleteSharedDomainURL(domainGuid))
}

func (repo CloudControllerDomainRepository) FirstOrDefault(orgGuid string, name *string) (domain models.DomainFields, error error) {
	if name == nil {
		domain, error = repo.defaultDomain(orgGuid)
	} else {
		domain, error = repo.FindByNameInOrg(*name, orgGuid)
	}
	return
}

func (repo CloudControllerDomainRepository) defaultDomain(orgGuid string) (models.DomainFields, error) {
	var foundDomain *models.DomainFields
	repo.ListDomainsForOrg(orgGuid, func(domain models.DomainFields) bool {
		foundDomain = &domain
		return !domain.Shared
	})

	if foundDomain == nil {
		return models.DomainFields{}, errors.New(T("Could not find a default domain"))
	}

	return *foundDomain, nil
}
