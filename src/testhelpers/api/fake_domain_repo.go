package api

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

	CreateSharedDomainDomain cf.Domain

	MapDomain cf.Domain
	MapSpace cf.Space
	MapApiResponse net.ApiResponse

	UnmapDomain cf.Domain
	UnmapSpace cf.Space
	UnmapApiResponse net.ApiResponse

	DeleteDomain cf.Domain
	DeleteApiResponse net.ApiResponse
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
		apiResponse = net.NewNotFoundApiResponse("%s %s not found","Domain", name)
	}
	if repo.FindByNameErr {
		apiResponse = net.NewApiResponseWithMessage("Error finding domain")
	}

	return
}

func (repo *FakeDomainRepository) FindByNameInOrg(name string, owningOrg cf.Organization) (domain cf.Domain, apiResponse net.ApiResponse) {

	domain = repo.FindByNameInOrgDomain
	apiResponse = repo.FindByNameInOrgApiResponse
	return
}

func (repo *FakeDomainRepository) Create(domainToCreate cf.Domain, owningOrg cf.Organization) (createdDomain cf.Domain, apiResponse net.ApiResponse){
	repo.ReserveDomainDomainToCreate = domainToCreate
	repo.ReserveDomainOwningOrg = owningOrg
	return
}

func (repo *FakeDomainRepository) CreateSharedDomain(domain cf.Domain) (apiResponse net.ApiResponse){
	repo.CreateSharedDomainDomain = domain
	return
}

func (repo *FakeDomainRepository) Delete(domain cf.Domain) (apiResponse net.ApiResponse) {
	repo.DeleteDomain = domain
	apiResponse = repo.DeleteApiResponse
	return
}

func (repo *FakeDomainRepository) Map(domain cf.Domain, space cf.Space) (apiResponse net.ApiResponse) {
	repo.MapDomain = domain
	repo.MapSpace = space
	apiResponse = repo.MapApiResponse
	return
}

func (repo *FakeDomainRepository) Unmap(domain cf.Domain, space cf.Space) (apiResponse net.ApiResponse) {
	repo.UnmapDomain = domain
	repo.UnmapSpace = space
	apiResponse = repo.UnmapApiResponse
	return
}
