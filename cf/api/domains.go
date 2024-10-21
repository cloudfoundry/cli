package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"

	. "code.cloudfoundry.org/cli/v7/cf/i18n"

	"code.cloudfoundry.org/cli/v7/cf/api/resources"
	"code.cloudfoundry.org/cli/v7/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/v7/cf/errors"
	"code.cloudfoundry.org/cli/v7/cf/models"
	"code.cloudfoundry.org/cli/v7/cf/net"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . DomainRepository

type DomainRepository interface {
	ListDomainsForOrg(orgGUID string, cb func(models.DomainFields) bool) error
	FindSharedByName(name string) (domain models.DomainFields, apiErr error)
	FindPrivateByName(name string) (domain models.DomainFields, apiErr error)
	FindByNameInOrg(name string, owningOrgGUID string) (domain models.DomainFields, apiErr error)
	Create(domainName string, owningOrgGUID string) (createdDomain models.DomainFields, apiErr error)
	CreateSharedDomain(domainName string, routerGroupGUID string) (apiErr error)
	Delete(domainGUID string) (apiErr error)
	DeleteSharedDomain(domainGUID string) (apiErr error)
	FirstOrDefault(orgGUID string, name *string) (domain models.DomainFields, error error)
}

type CloudControllerDomainRepository struct {
	config  coreconfig.Reader
	gateway net.Gateway
}

func NewCloudControllerDomainRepository(config coreconfig.Reader, gateway net.Gateway) CloudControllerDomainRepository {
	return CloudControllerDomainRepository{
		config:  config,
		gateway: gateway,
	}
}

func (repo CloudControllerDomainRepository) ListDomainsForOrg(orgGUID string, cb func(models.DomainFields) bool) error {
	path := fmt.Sprintf("/v2/organizations/%s/private_domains", orgGUID)
	err := repo.listDomains(path, cb)
	if err != nil {
		return err
	}
	err = repo.listDomains("/v2/shared_domains?inline-relations-depth=1", cb)
	return err
}

func (repo CloudControllerDomainRepository) listDomains(path string, cb func(models.DomainFields) bool) error {
	return repo.gateway.ListPaginatedResources(
		repo.config.APIEndpoint(),
		path,
		resources.DomainResource{},
		func(resource interface{}) bool {
			return cb(resource.(resources.DomainResource).ToFields())
		})
}

func (repo CloudControllerDomainRepository) isOrgDomain(orgGUID string, domain models.DomainFields) bool {
	return orgGUID == domain.OwningOrganizationGUID || domain.Shared
}

func (repo CloudControllerDomainRepository) FindSharedByName(name string) (domain models.DomainFields, apiErr error) {
	path := fmt.Sprintf("/v2/shared_domains?inline-relations-depth=1&q=name:%s", url.QueryEscape(name))
	return repo.findOneWithPath(path, name)
}

func (repo CloudControllerDomainRepository) FindPrivateByName(name string) (domain models.DomainFields, apiErr error) {
	path := fmt.Sprintf("/v2/private_domains?inline-relations-depth=1&q=name:%s", url.QueryEscape(name))
	return repo.findOneWithPath(path, name)
}

func (repo CloudControllerDomainRepository) FindByNameInOrg(name string, orgGUID string) (models.DomainFields, error) {
	path := fmt.Sprintf("/v2/organizations/%s/private_domains?inline-relations-depth=1&q=name:%s", orgGUID, url.QueryEscape(name))
	domain, err := repo.findOneWithPath(path, name)

	switch err.(type) {
	case *errors.ModelNotFoundError:
		domain, err = repo.FindSharedByName(name)
		if err != nil {
			return models.DomainFields{}, err
		}
		if !domain.Shared {
			err = errors.NewModelNotFoundError("Domain", name)
		}
	}

	return domain, err
}

func (repo CloudControllerDomainRepository) findOneWithPath(path, name string) (models.DomainFields, error) {
	var domain models.DomainFields

	foundDomain := false
	err := repo.listDomains(path, func(result models.DomainFields) bool {
		domain = result
		foundDomain = true
		return false
	})

	if err == nil && !foundDomain {
		err = errors.NewModelNotFoundError("Domain", name)
	}

	return domain, err
}

func (repo CloudControllerDomainRepository) Create(domainName string, owningOrgGUID string) (createdDomain models.DomainFields, err error) {
	data, err := json.Marshal(resources.DomainEntity{
		Name:                   domainName,
		OwningOrganizationGUID: owningOrgGUID,
		Wildcard:               true,
	})

	if err != nil {
		return
	}

	resource := new(resources.DomainResource)
	err = repo.gateway.CreateResource(
		repo.config.APIEndpoint(),
		"/v2/private_domains",
		bytes.NewReader(data),
		resource)

	if err != nil {
		return
	}

	createdDomain = resource.ToFields()
	return
}

func (repo CloudControllerDomainRepository) CreateSharedDomain(domainName string, routerGroupGUID string) error {
	data, err := json.Marshal(resources.DomainEntity{
		Name:            domainName,
		RouterGroupGUID: routerGroupGUID,
		Wildcard:        true,
	})
	if err != nil {
		return err
	}

	return repo.gateway.CreateResource(
		repo.config.APIEndpoint(),
		"/v2/shared_domains",
		bytes.NewReader(data),
	)
}

func (repo CloudControllerDomainRepository) Delete(domainGUID string) error {
	path := fmt.Sprintf("/v2/private_domains/%s?recursive=true", domainGUID)
	return repo.gateway.DeleteResource(
		repo.config.APIEndpoint(),
		path)
}

func (repo CloudControllerDomainRepository) DeleteSharedDomain(domainGUID string) error {
	path := fmt.Sprintf("/v2/shared_domains/%s?recursive=true", domainGUID)
	return repo.gateway.DeleteResource(
		repo.config.APIEndpoint(),
		path)
}

func (repo CloudControllerDomainRepository) FirstOrDefault(orgGUID string, name *string) (domain models.DomainFields, error error) {
	if name == nil {
		domain, error = repo.defaultDomain(orgGUID)
	} else {
		domain, error = repo.FindByNameInOrg(*name, orgGUID)
	}
	return
}

func (repo CloudControllerDomainRepository) defaultDomain(orgGUID string) (models.DomainFields, error) {
	var foundDomain *models.DomainFields
	err := repo.ListDomainsForOrg(orgGUID, func(domain models.DomainFields) bool {
		foundDomain = &domain
		return !domain.Shared
	})
	if err != nil {
		return models.DomainFields{}, err
	}

	if foundDomain == nil {
		return models.DomainFields{}, errors.New(T("Could not find a default domain"))
	}

	return *foundDomain, nil
}
