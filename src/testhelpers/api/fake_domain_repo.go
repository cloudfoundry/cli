package api

import (
	"cf/models"
	"cf/net"
)

type FakeDomainRepository struct {
	ListDomainsForOrgDomainsGuid string
	ListDomainsForOrgDomains     []models.DomainFields
	ListDomainsForOrgApiResponse net.ApiResponse

	ListSharedDomainsDomains     []models.DomainFields
	ListSharedDomainsApiResponse net.ApiResponse

	ListDomainsDomains     []models.DomainFields
	ListDomainsApiResponse net.ApiResponse

	FindByNameInOrgName        string
	FindByNameInOrgGuid        string
	FindByNameInOrgDomain      models.DomainFields
	FindByNameInOrgApiResponse net.ApiResponse

	FindByNameName     string
	FindByNameDomain   models.DomainFields
	FindByNameNotFound bool
	FindByNameErr      bool

	CreateDomainName          string
	CreateDomainOwningOrgGuid string

	CreateSharedDomainName string

	DeleteDomainGuid  string
	DeleteApiResponse net.ApiResponse

	DeleteSharedDomainGuid  string
	DeleteSharedApiResponse net.ApiResponse
}

func (repo *FakeDomainRepository) ListDomainsForOrg(orgGuid string, cb func(models.DomainFields) bool) net.ApiResponse {
	repo.ListDomainsForOrgDomainsGuid = orgGuid
	for _, d := range repo.ListDomainsForOrgDomains {
		cb(d)
	}
	return repo.ListDomainsForOrgApiResponse
}

func (repo *FakeDomainRepository) ListSharedDomains(cb func(models.DomainFields) bool) net.ApiResponse {
	for _, d := range repo.ListSharedDomainsDomains {
		cb(d)
	}
	return repo.ListSharedDomainsApiResponse
}

func (repo *FakeDomainRepository) ListDomains(cb func(models.DomainFields) bool) net.ApiResponse {
	for _, d := range repo.ListDomainsDomains {
		cb(d)
	}
	return repo.ListDomainsApiResponse
}

func (repo *FakeDomainRepository) FindByName(name string) (domain models.DomainFields, apiResponse net.ApiResponse) {
	repo.FindByNameName = name
	domain = repo.FindByNameDomain

	if repo.FindByNameNotFound {
		apiResponse = net.NewNotFoundApiResponse("%s %s not found", "Domain", name)
	}
	if repo.FindByNameErr {
		apiResponse = net.NewApiResponseWithMessage("Error finding domain")
	}

	return
}

func (repo *FakeDomainRepository) FindByNameInOrg(name string, owningOrgGuid string) (domain models.DomainFields, apiResponse net.ApiResponse) {
	repo.FindByNameInOrgName = name
	repo.FindByNameInOrgGuid = owningOrgGuid
	domain = repo.FindByNameInOrgDomain
	apiResponse = repo.FindByNameInOrgApiResponse
	return
}

func (repo *FakeDomainRepository) Create(domainName string, owningOrgGuid string) (createdDomain models.DomainFields, apiResponse net.ApiResponse) {
	repo.CreateDomainName = domainName
	repo.CreateDomainOwningOrgGuid = owningOrgGuid
	return
}

func (repo *FakeDomainRepository) CreateSharedDomain(domainName string) (apiResponse net.ApiResponse) {
	repo.CreateSharedDomainName = domainName
	return
}

func (repo *FakeDomainRepository) Delete(domainGuid string) (apiResponse net.ApiResponse) {
	repo.DeleteDomainGuid = domainGuid
	apiResponse = repo.DeleteApiResponse
	return
}

func (repo *FakeDomainRepository) DeleteSharedDomain(domainGuid string) (apiResponse net.ApiResponse) {
	repo.DeleteSharedDomainGuid = domainGuid
	apiResponse = repo.DeleteSharedApiResponse
	return
}
