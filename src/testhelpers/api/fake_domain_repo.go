package api

import (
	"cf"
	"cf/net"
)

type FakeDomainRepository struct {
	DefaultAppDomain cf.Domain

	FindAllInCurrentSpaceDomains []cf.Domain

	FindAllByOrgOrgGuid string
	FindAllByOrgDomains []cf.Domain

	FindByNameInOrgDomain cf.Domain
	FindByNameInOrgApiResponse net.ApiResponse

	FindByNameInCurrentSpaceName string

	FindByNameName string
	FindByNameDomain cf.Domain
	FindByNameNotFound bool
	FindByNameErr bool

	CreateDomainName string
	CreateDomainOwningOrgGuid string

	CreateSharedDomainName string

	MapDomainGuid string
	MapSpaceGuid string
	MapApiResponse net.ApiResponse

	UnmapDomainGuid string
	UnmapSpaceGuid string
	UnmapApiResponse net.ApiResponse

	DeleteDomainGuid string
	DeleteApiResponse net.ApiResponse
}

func (repo *FakeDomainRepository) FindDefaultAppDomain() (domain cf.Domain, apiResponse net.ApiResponse){
	domain = repo.DefaultAppDomain
	return
}

func (repo *FakeDomainRepository) FindAllByOrg(orgGuid string)(domains []cf.Domain, apiResponse net.ApiResponse){
	repo.FindAllByOrgOrgGuid = orgGuid
	domains = repo.FindAllByOrgDomains

	return
}

func (repo *FakeDomainRepository) FindByName(name string) (domain cf.Domain, apiResponse net.ApiResponse){
	repo.FindByNameName = name
	domain = repo.FindByNameDomain

	if repo.FindByNameNotFound {
		apiResponse = net.NewNotFoundApiResponse("%s %s not found","Domain", name)
	}
	if repo.FindByNameErr {
		apiResponse = net.NewApiResponseWithMessage("Error finding domain")
	}

	return
}


func (repo *FakeDomainRepository) FindByNameInCurrentSpace(name string) (domain cf.Domain, apiResponse net.ApiResponse){
	repo.FindByNameInCurrentSpaceName = name
	domain = repo.FindByNameDomain

	if repo.FindByNameNotFound {
		apiResponse = net.NewNotFoundApiResponse("%s %s not found","Domain", name)
	}
	if repo.FindByNameErr {
		apiResponse = net.NewApiResponseWithMessage("Error finding domain")
	}

	return
}

func (repo *FakeDomainRepository) FindByNameInOrg(name string, owningOrgGuid string) (domain cf.Domain, apiResponse net.ApiResponse) {
	domain = repo.FindByNameInOrgDomain
	apiResponse = repo.FindByNameInOrgApiResponse
	return
}

func (repo *FakeDomainRepository) Create(domainName string, owningOrgGuid string) (createdDomain cf.DomainFields, apiResponse net.ApiResponse){
	repo.CreateDomainName = domainName
	repo.CreateDomainOwningOrgGuid = owningOrgGuid
	return
}

func (repo *FakeDomainRepository) CreateSharedDomain(domainName string) (apiResponse net.ApiResponse){
	repo.CreateSharedDomainName = domainName
	return
}

func (repo *FakeDomainRepository) Delete(domainGuid string) (apiResponse net.ApiResponse) {
	repo.DeleteDomainGuid = domainGuid
	apiResponse = repo.DeleteApiResponse
	return
}

func (repo *FakeDomainRepository) Map(domainGuid , spaceGuid string) (apiResponse net.ApiResponse) {
	repo.MapDomainGuid = domainGuid
	repo.MapSpaceGuid = spaceGuid
	apiResponse = repo.MapApiResponse
	return
}

func (repo *FakeDomainRepository) Unmap(domainGuid, spaceGuid string) (apiResponse net.ApiResponse) {
	repo.UnmapDomainGuid = domainGuid
	repo.UnmapSpaceGuid = spaceGuid
	apiResponse = repo.UnmapApiResponse
	return
}
