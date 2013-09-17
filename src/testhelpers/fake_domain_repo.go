package testhelpers

import (
	"cf"
	"cf/net"
)

type FakeDomainRepository struct {
	FindAllDomains []cf.Domain

	FindByNameName string
	FindByNameDomain cf.Domain
}

func (repo *FakeDomainRepository) FindAll() (domains []cf.Domain, apiErr *net.ApiError){
	return repo.FindAllDomains, nil
}

func (repo *FakeDomainRepository) FindByName(name string) (domain cf.Domain, apiErr *net.ApiError){
	repo.FindByNameName = name
	return repo.FindByNameDomain, nil
}
