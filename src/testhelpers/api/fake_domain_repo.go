package api

import (
	"cf/errors"
	"cf/models"
)

type FakeDomainRepository struct {
	ListDomainsForOrgDomainsGuid string
	ListDomainsForOrgDomains     []models.DomainFields
	ListDomainsForOrgApiResponse error

	ListSharedDomainsDomains     []models.DomainFields
	ListSharedDomainsApiResponse error

	ListDomainsDomains     []models.DomainFields
	ListDomainsApiResponse error

	FindByNameInOrgName        string
	FindByNameInOrgGuid        string
	FindByNameInOrgDomain      models.DomainFields
	FindByNameInOrgApiResponse error

	FindByNameName     string
	FindByNameDomain   models.DomainFields
	FindByNameNotFound bool
	FindByNameErr      bool

	CreateDomainName          string
	CreateDomainOwningOrgGuid string

	CreateSharedDomainName string

	DeleteDomainGuid  string
	DeleteApiResponse error

	DeleteSharedDomainGuid  string
	DeleteSharedApiResponse error
}

func (repo *FakeDomainRepository) ListDomainsForOrg(orgGuid string, cb func(models.DomainFields) bool) error {
	repo.ListDomainsForOrgDomainsGuid = orgGuid
	for _, d := range repo.ListDomainsForOrgDomains {
		cb(d)
	}
	return repo.ListDomainsForOrgApiResponse
}

func (repo *FakeDomainRepository) ListSharedDomains(cb func(models.DomainFields) bool) error {
	for _, d := range repo.ListSharedDomainsDomains {
		cb(d)
	}
	return repo.ListSharedDomainsApiResponse
}

func (repo *FakeDomainRepository) ListDomains(cb func(models.DomainFields) bool) error {
	for _, d := range repo.ListDomainsDomains {
		cb(d)
	}
	return repo.ListDomainsApiResponse
}

func (repo *FakeDomainRepository) FindByName(name string) (domain models.DomainFields, apiErr error) {
	repo.FindByNameName = name
	domain = repo.FindByNameDomain

	if repo.FindByNameNotFound {
		apiErr = errors.NewModelNotFoundError("Domain", name)
	}
	if repo.FindByNameErr {
		apiErr = errors.New("Error finding domain")
	}

	return
}

func (repo *FakeDomainRepository) FindByNameInOrg(name string, owningOrgGuid string) (domain models.DomainFields, apiErr error) {
	repo.FindByNameInOrgName = name
	repo.FindByNameInOrgGuid = owningOrgGuid
	domain = repo.FindByNameInOrgDomain
	apiErr = repo.FindByNameInOrgApiResponse
	return
}

func (repo *FakeDomainRepository) Create(domainName string, owningOrgGuid string) (createdDomain models.DomainFields, apiErr error) {
	repo.CreateDomainName = domainName
	repo.CreateDomainOwningOrgGuid = owningOrgGuid
	return
}

func (repo *FakeDomainRepository) CreateSharedDomain(domainName string) (apiErr error) {
	repo.CreateSharedDomainName = domainName
	return
}

func (repo *FakeDomainRepository) Delete(domainGuid string) (apiErr error) {
	repo.DeleteDomainGuid = domainGuid
	apiErr = repo.DeleteApiResponse
	return
}

func (repo *FakeDomainRepository) DeleteSharedDomain(domainGuid string) (apiErr error) {
	repo.DeleteSharedDomainGuid = domainGuid
	apiErr = repo.DeleteSharedApiResponse
	return
}
