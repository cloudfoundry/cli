package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"fmt"
	"strings"
)

type DomainRepository interface {
	FindAll() (domains []cf.Domain, apiErr *net.ApiError)
	FindByName(name string) (domain cf.Domain, apiErr *net.ApiError)
}

type CloudControllerDomainRepository struct {
	config  *configuration.Configuration
	gateway net.Gateway
}

func NewCloudControllerDomainRepository(config *configuration.Configuration, gateway net.Gateway) (repo CloudControllerDomainRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (repo CloudControllerDomainRepository) FindAll() (domains []cf.Domain, apiErr *net.ApiError) {
	path := fmt.Sprintf("%s/v2/spaces/%s/domains", repo.config.Target, repo.config.Space.Guid)
	request, apiErr := repo.gateway.NewRequest("GET", path, repo.config.AccessToken, nil)
	if apiErr != nil {
		return
	}

	response := new(ApiResponse)
	apiErr = repo.gateway.PerformRequestForJSONResponse(request, response)
	if apiErr != nil {
		return
	}

	for _, r := range response.Resources {
		domains = append(domains, cf.Domain{r.Entity.Name, r.Metadata.Guid})
	}

	return
}

func (repo CloudControllerDomainRepository) FindByName(name string) (domain cf.Domain, apiErr *net.ApiError) {
	domains, apiErr := repo.FindAll()

	if apiErr != nil {
		return
	}

	if name == "" {
		domain = domains[0]
	} else {
		apiErr = net.NewApiErrorWithMessage("Could not find domain with name %s", name)

		for _, d := range domains {
			if d.Name == strings.ToLower(name) {
				domain = d
				apiErr = nil
			}
		}
	}

	return
}
