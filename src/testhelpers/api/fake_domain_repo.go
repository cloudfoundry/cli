package api

import (
"cf/models"
	"cf/net"
	"cf/api"
)

type FakeDomainRepository struct {
	FindAllInCurrentSpaceDomains []models.Domain

	ListDomainsForOrgDomainsGuid string
	ListDomainsForOrgDomains     []models.Domain
	ListDomainsForOrgApiResponse net.ApiResponse

	ListSharedDomainsDomains     []models.Domain
	ListSharedDomainsApiResponse net.ApiResponse

	ListDomainsDomains     []models.Domain
	ListDomainsApiResponse net.ApiResponse

	FindByNameInOrgDomain      models.Domain
	FindByNameInOrgApiResponse net.ApiResponse

	FindByNameInCurrentSpaceName string

	FindByNameName     string
	FindByNameDomain   models.Domain
	FindByNameNotFound bool
	FindByNameErr      bool

	CreateDomainName          string
	CreateDomainOwningOrgGuid string

	CreateSharedDomainName string

	DeleteDomainGuid string
	DeleteApiResponse net.ApiResponse

	DeleteSharedDomainGuid string
	DeleteSharedApiResponse net.ApiResponse
}

func (repo *FakeDomainRepository) ListDomainsForOrg(orgGuid string, cb api.ListDomainsCallback) net.ApiResponse {
	repo.ListDomainsForOrgDomainsGuid = orgGuid
	if len(repo.ListDomainsForOrgDomains) > 0 {
		cb(repo.ListDomainsForOrgDomains)
	}
	return repo.ListDomainsForOrgApiResponse
}

func (repo *FakeDomainRepository) ListSharedDomains(cb api.ListDomainsCallback) net.ApiResponse {
	if len(repo.ListSharedDomainsDomains) > 0 {
		cb(repo.ListSharedDomainsDomains)
	}
	return repo.ListSharedDomainsApiResponse
}

func (repo *FakeDomainRepository) ListDomains(cb api.ListDomainsCallback) net.ApiResponse {
	if len(repo.ListDomainsDomains) > 0 {
		cb(repo.ListDomainsDomains)
	}
	return repo.ListDomainsApiResponse
}

func (repo *FakeDomainRepository) FindByName(name string) (domain models.Domain, apiResponse net.ApiResponse) {
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

func (repo *FakeDomainRepository) FindByNameInCurrentSpace(name string) (domain models.Domain, apiResponse net.ApiResponse) {
	repo.FindByNameInCurrentSpaceName = name
	domain = repo.FindByNameDomain

	if repo.FindByNameNotFound {
		apiResponse = net.NewNotFoundApiResponse("%s %s not found", "Domain", name)
	}
	if repo.FindByNameErr {
		apiResponse = net.NewApiResponseWithMessage("Error finding domain")
	}

	return
}

func (repo *FakeDomainRepository) FindByNameInOrg(name string, owningOrgGuid string) (domain models.Domain, apiResponse net.ApiResponse) {
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
