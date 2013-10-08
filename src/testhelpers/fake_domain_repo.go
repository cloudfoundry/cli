package testhelpers

import (
	"cf"
	"cf/net"
)

type FakeDomainRepository struct {
	FindAllInCurrentSpaceDomains []cf.Domain

	FindAllByOrgOrg cf.Organization
	FindAllByOrgDomains []cf.Domain

	FindByNameInOrgDomain cf.Domain
	FindByNameInOrgApiResponse net.ApiResponse

	FindByNameName string
	FindByNameDomain cf.Domain
	FindByNameNotFound bool
	FindByNameErr bool

	ReserveDomainDomainToCreate cf.Domain
	ReserveDomainOwningOrg cf.Organization

	MapDomainDomain cf.Domain
	MapDomainSpace cf.Space
	MapDomainApiResponse net.ApiResponse

	UnmapDomainDomain cf.Domain
	UnmapDomainSpace cf.Space
	UnmapDomainApiResponse net.ApiResponse

	DeleteDomainDomain cf.Domain
	DeleteDomainApiResponse net.ApiResponse
}

func (repo *FakeDomainRepository) FindAllInCurrentSpace() (domains []cf.Domain, apiResponse net.ApiResponse){
	domains = repo.FindAllInCurrentSpaceDomains
	return
}

func (repo *FakeDomainRepository) FindAllByOrg(org cf.Organization)(domains []cf.Domain, apiResponse net.ApiResponse){
	repo.FindAllByOrgOrg = org
	domains = repo.FindAllByOrgDomains

	return
}

func (repo *FakeDomainRepository) FindByNameInCurrentSpace(name string) (domain cf.Domain, apiResponse net.ApiResponse){
	repo.FindByNameName = name
	domain = repo.FindByNameDomain

	if repo.FindByNameNotFound {
		apiResponse = net.NewNotFoundApiResponse("Domain", name)
	}
	if repo.FindByNameErr {
		apiResponse = net.NewApiResponseWithMessage("Error finding domain")
	}

	return
}

func (repo *FakeDomainRepository) Create(domainToCreate cf.Domain, owningOrg cf.Organization) (createdDomain cf.Domain, apiResponse net.ApiResponse){
	repo.ReserveDomainDomainToCreate = domainToCreate
	repo.ReserveDomainOwningOrg = owningOrg
	return
}

func (repo *FakeDomainRepository) MapDomain(domain cf.Domain, space cf.Space) (apiResponse net.ApiResponse) {
	repo.MapDomainDomain = domain
	repo.MapDomainSpace = space
	apiResponse = repo.MapDomainApiResponse
	return
}

func (repo *FakeDomainRepository) UnmapDomain(domain cf.Domain, space cf.Space) (apiResponse net.ApiResponse) {
	repo.UnmapDomainDomain = domain
	repo.UnmapDomainSpace = space
	apiResponse = repo.UnmapDomainApiResponse
	return
}

func (repo *FakeDomainRepository) FindByNameInOrg(name string, owningOrg cf.Organization) (domain cf.Domain, apiResponse net.ApiResponse) {

	domain = repo.FindByNameInOrgDomain
	apiResponse = repo.FindByNameInOrgApiResponse
	return
}

func (repo *FakeDomainRepository) DeleteDomain(domain cf.Domain) (apiResponse net.ApiResponse) {
	repo.DeleteDomainDomain = domain
	apiResponse = repo.DeleteDomainApiResponse
	return
}
