package api

import (
	"cf"
	"cf/configuration"
	"errors"
	"fmt"
	"strings"
)

type DomainRepository interface {
	FindAll(config *configuration.Configuration) (domains []cf.Domain, err error)
	FindByName(config *configuration.Configuration, name string) (domain cf.Domain, err error)
}

type CloudControllerDomainRepository struct {
}

func (repo CloudControllerDomainRepository) FindAll(config *configuration.Configuration) (domains []cf.Domain, err error) {
	path := fmt.Sprintf("%s/v2/spaces/%s/domains", config.Target, config.Space.Guid)
	request, err := NewAuthorizedRequest("GET", path, config.AccessToken, nil)
	if err != nil {
		return
	}

	response := new(ApiResponse)
	err = PerformRequestAndParseResponse(request, response)
	if err != nil {
		return
	}

	for _, r := range response.Resources {
		domains = append(domains, cf.Domain{r.Entity.Name, r.Metadata.Guid})
	}

	return
}

func (repo CloudControllerDomainRepository) FindByName(config *configuration.Configuration, name string) (domain cf.Domain, err error) {
	domains, err := repo.FindAll(config)

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
