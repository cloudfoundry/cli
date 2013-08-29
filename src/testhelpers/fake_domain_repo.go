package testhelpers

import (
	"cf"
	"cf/configuration"
)

type FakeDomainRepository struct {
	FindAllDomains []cf.Domain

	FindByNameName string
	FindByNameDomain cf.Domain
}

func (repo *FakeDomainRepository) FindAll(config *configuration.Configuration) (domains []cf.Domain, err error){
	return repo.FindAllDomains, nil
}

func (repo *FakeDomainRepository) FindByName(config *configuration.Configuration, name string) (domain cf.Domain, err error){
	repo.FindByNameName = name
	return repo.FindByNameDomain, nil
}

