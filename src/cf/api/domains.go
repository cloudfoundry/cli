package api

import (
	"cf"
	"cf/configuration"
	"errors"
	"fmt"
	"strings"
)

type DomainRepository interface {
	FindAll() (domains []cf.Domain, err error)
	FindByName(name string) (domain cf.Domain, err error)
}

type CloudControllerDomainRepository struct {
	config *configuration.Configuration
}

func NewCloudControllerDomainRepository(config *configuration.Configuration) (repo CloudControllerDomainRepository) {
	repo.config = config
	return
}

func (repo CloudControllerDomainRepository) FindAll() (domains []cf.Domain, err error) {
	path := fmt.Sprintf("%s/v2/spaces/%s/domains", repo.config.Target, repo.config.Space.Guid)
	request, err := NewRequest("GET", path, repo.config.AccessToken, nil)
	if err != nil {
		return
	}

	response := new(ApiResponse)
	_, err = PerformRequestAndParseResponse(request, response)
	if err != nil {
		return
	}

	for _, r := range response.Resources {
		domains = append(domains, cf.Domain{r.Entity.Name, r.Metadata.Guid})
	}

	return
}

func (repo CloudControllerDomainRepository) FindByName(name string) (domain cf.Domain, err error) {
	domains, err := repo.FindAll()

	if err != nil {
		return
	}

	if name == "" {
		domain = domains[0]
	} else {
		err = errors.New(fmt.Sprintf("Could not find domain with name %s", name))

		for _, d := range domains {
			if d.Name == strings.ToLower(name) {
				domain = d
				err = nil
			}
		}
	}

	return
}
