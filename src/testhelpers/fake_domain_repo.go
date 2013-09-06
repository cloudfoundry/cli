package testhelpers

import (
	"cf"
)

type FakeDomainRepository struct {
	FindAllDomains []cf.Domain

	FindByNameName string
	FindByNameDomain cf.Domain
}

func (repo *FakeDomainRepository) FindAll() (domains []cf.Domain, err error){
	return repo.FindAllDomains, nil
}

func (repo *FakeDomainRepository) FindByName(name string) (domain cf.Domain, err error){
	repo.FindByNameName = name
	return repo.FindByNameDomain, nil
}

