package testhelpers

import (
	"cf"
	"cf/net"
)

type FakeDomainRepository struct {
	FindAllDomains []cf.Domain

	FindAllByOrgOrg cf.Organization
	FindAllByOrgDomains []cf.Domain

	FindByNameName string
	FindByNameDomain cf.Domain

	ParkDomainDomainToCreate cf.Domain
	ParkDomainOwningOrg cf.Organization
}

func (repo *FakeDomainRepository) FindAll() (domains []cf.Domain, apiStatus net.ApiStatus){
	domains = repo.FindAllDomains
	return
}

func (repo *FakeDomainRepository) FindAllByOrg(org cf.Organization)(domains []cf.Domain, apiStatus net.ApiStatus){
	repo.FindAllByOrgOrg = org
	domains = repo.FindAllByOrgDomains

	return
}

func (repo *FakeDomainRepository) FindByName(name string) (domain cf.Domain, apiStatus net.ApiStatus){
	repo.FindByNameName = name
	domain = repo.FindByNameDomain
	return
}

func (repo *FakeDomainRepository) Create(domainToCreate cf.Domain, owningOrg cf.Organization) (createdDomain cf.Domain, apiStatus net.ApiStatus){
	repo.ParkDomainDomainToCreate = domainToCreate
	repo.ParkDomainOwningOrg = owningOrg
	return
}
