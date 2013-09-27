package testhelpers

import (
	"cf"
	"cf/net"
)

type FakeDomainRepository struct {
	FindAllDomains []cf.Domain

	FindByNameName string
	FindByNameDomain cf.Domain

	CreateDomainName string
}

func (repo *FakeDomainRepository) FindAll() (domains []cf.Domain, apiErr *net.ApiError){
	return repo.FindAllDomains, nil
}

func (repo *FakeDomainRepository) FindByName(name string) (domain cf.Domain, apiErr *net.ApiError){
	repo.FindByNameName = name
	return repo.FindByNameDomain, nil
}

func (repo *FakeDomainRepository) Create(domainToCreate cf.Domain) (createdDomain cf.Domain, apiErr *net.ApiError){
	repo.CreateDomainName = domainToCreate.Name
	return
}
