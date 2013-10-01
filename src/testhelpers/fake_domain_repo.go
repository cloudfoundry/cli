package testhelpers

import (
	"cf"
	"cf/net"
)

type FakeDomainRepository struct {
	FindAllDomains []cf.Domain

	FindByNameName string
	FindByNameDomain cf.Domain

	CreateDomainDomainToCreate cf.Domain
	CreateDomainOwningOrg cf.Organization
}

func (repo *FakeDomainRepository) FindAll() (domains []cf.Domain, apiStatus net.ApiStatus){
	domains = repo.FindAllDomains
	return
}

func (repo *FakeDomainRepository) FindByName(name string) (domain cf.Domain, apiStatus net.ApiStatus){
	repo.FindByNameName = name
	domain = repo.FindByNameDomain
	return
}

func (repo *FakeDomainRepository) Create(domainToCreate cf.Domain, owningOrg cf.Organization) (createdDomain cf.Domain, apiStatus net.ApiStatus){
	repo.CreateDomainDomainToCreate = domainToCreate
	repo.CreateDomainOwningOrg = owningOrg
	return
}
