package testhelpers

import (
	"cf"
	"cf/configuration"
)

type FakeDomainRepository struct {
	Domains []cf.Domain
}

func (repo *FakeDomainRepository) FindAll(config *configuration.Configuration) (domains []cf.Domain, err error){
	return repo.Domains, nil
}

