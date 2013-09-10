package api

import (
	"cf"
	"cf/configuration"
	"fmt"
	"strings"
)

type DomainRepository interface {
	FindAll() (domains []cf.Domain, apiErr *ApiError)
	FindByName(name string) (domain cf.Domain, apiErr *ApiError)
}

type CloudControllerDomainRepository struct {
	config    *configuration.Configuration
	apiClient ApiClient
}

func NewCloudControllerDomainRepository(config *configuration.Configuration, apiClient ApiClient) (repo CloudControllerDomainRepository) {
	repo.config = config
	repo.apiClient = apiClient
	return
}

func (repo CloudControllerDomainRepository) FindAll() (domains []cf.Domain, apiErr *ApiError) {
	path := fmt.Sprintf("%s/v2/spaces/%s/domains", repo.config.Target, repo.config.Space.Guid)
	request, apiErr := NewRequest("GET", path, repo.config.AccessToken, nil)
	if apiErr != nil {
		return
	}

	response := new(ApiResponse)
	apiErr = repo.apiClient.PerformRequestAndParseResponse(request, response)
	if apiErr != nil {
		return
	}

	for _, r := range response.Resources {
		domains = append(domains, cf.Domain{r.Entity.Name, r.Metadata.Guid})
	}

	return
}

func (repo CloudControllerDomainRepository) FindByName(name string) (domain cf.Domain, apiErr *ApiError) {
	domains, apiErr := repo.FindAll()

	if apiErr != nil {
		return
	}

	if name == "" {
		domain = domains[0]
	} else {
		apiErr = NewApiErrorWithMessage("Could not find domain with name %s", name)

		for _, d := range domains {
			if d.Name == strings.ToLower(name) {
				domain = d
				apiErr = nil
			}
		}
	}

	return
}
