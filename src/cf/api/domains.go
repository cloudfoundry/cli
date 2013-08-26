package api

import (
	"cf"
	"cf/configuration"
	"fmt"
)

type DomainRepository interface {
	FindAll(config *configuration.Configuration) (domains []cf.Domain, err error)
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
	err = PerformRequestForBody(request, response)
	if err != nil {
		return
	}

	for _, r := range response.Resources {
		domains = append(domains, cf.Domain{r.Entity.Name, r.Metadata.Guid})
	}

	return
}
