package testhelpers

import (
	"cf"
	"cf/api"
)

type FakeDomainRepository struct {
	FindAllDomains []cf.Domain

	FindByNameName string
	FindByNameDomain cf.Domain
}

func (repo *FakeDomainRepository) FindAll() (domains []cf.Domain, apiErr *api.ApiError){
	return repo.FindAllDomains, nil
}

func (repo *FakeDomainRepository) FindByName(name string) (domain cf.Domain, apiErr *api.ApiError){
	repo.FindByNameName = name
	return repo.FindByNameDomain, nil
}
