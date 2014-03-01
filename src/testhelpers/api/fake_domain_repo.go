package api

import (
	"cf/errors"
	"cf/models"
)

type FakeDomainRepository struct {
	ListDomainsForOrgDomainsGuid string
	ListDomainsForOrgDomains     []models.DomainFields
	ListDomainsForOrgApiResponse errors.Error

	ListSharedDomainsDomains     []models.DomainFields
	ListSharedDomainsApiResponse errors.Error

	ListDomainsDomains     []models.DomainFields
	ListDomainsApiResponse errors.Error

	FindByNameInOrgName        string
	FindByNameInOrgGuid        string
	FindByNameInOrgDomain      models.DomainFields
	FindByNameInOrgApiResponse errors.Error

	FindByNameName     string
	FindByNameDomain   models.DomainFields
	FindByNameNotFound bool
	FindByNameErr      bool

	CreateDomainName          string
	CreateDomainOwningOrgGuid string

	CreateSharedDomainName string

	DeleteDomainGuid  string
	DeleteApiResponse errors.Error

	DeleteSharedDomainGuid  string
	DeleteSharedApiResponse errors.Error
}

func (repo *FakeDomainRepository) ListDomainsForOrg(orgGuid string, cb func(models.DomainFields) bool) errors.Error {
	repo.ListDomainsForOrgDomainsGuid = orgGuid
	for _, d := range repo.ListDomainsForOrgDomains {
		cb(d)
	}
	return repo.ListDomainsForOrgApiResponse
}

func (repo *FakeDomainRepository) ListSharedDomains(cb func(models.DomainFields) bool) errors.Error {
	for _, d := range repo.ListSharedDomainsDomains {
		cb(d)
	}
	return repo.ListSharedDomainsApiResponse
}

func (repo *FakeDomainRepository) ListDomains(cb func(models.DomainFields) bool) errors.Error {
	for _, d := range repo.ListDomainsDomains {
		cb(d)
	}
	return repo.ListDomainsApiResponse
}

func (repo *FakeDomainRepository) FindByName(name string) (domain models.DomainFields, apiResponse errors.Error) {
	repo.FindByNameName = name
	domain = repo.FindByNameDomain

	if repo.FindByNameNotFound {
		apiResponse = errors.NewNotFoundError("%s %s not found", "Domain", name)
	}
	if repo.FindByNameErr {
		apiResponse = errors.NewErrorWithMessage("Error finding domain")
	}

	return
}

func (repo *FakeDomainRepository) FindByNameInOrg(name string, owningOrgGuid string) (domain models.DomainFields, apiResponse errors.Error) {
	repo.FindByNameInOrgName = name
	repo.FindByNameInOrgGuid = owningOrgGuid
	domain = repo.FindByNameInOrgDomain
	apiResponse = repo.FindByNameInOrgApiResponse
	return
}

func (repo *FakeDomainRepository) Create(domainName string, owningOrgGuid string) (createdDomain models.DomainFields, apiResponse errors.Error) {
	repo.CreateDomainName = domainName
	repo.CreateDomainOwningOrgGuid = owningOrgGuid
	return
}

func (repo *FakeDomainRepository) CreateSharedDomain(domainName string) (apiResponse errors.Error) {
	repo.CreateSharedDomainName = domainName
	return
}

func (repo *FakeDomainRepository) Delete(domainGuid string) (apiResponse errors.Error) {
	repo.DeleteDomainGuid = domainGuid
	apiResponse = repo.DeleteApiResponse
	return
}

func (repo *FakeDomainRepository) DeleteSharedDomain(domainGuid string) (apiResponse errors.Error) {
	repo.DeleteSharedDomainGuid = domainGuid
	apiResponse = repo.DeleteSharedApiResponse
	return
}
