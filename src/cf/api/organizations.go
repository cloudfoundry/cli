package api

import (
	"cf"
	"cf/configuration"
	"errors"
	"strings"
)

type OrganizationRepository interface {
	FindAll() (orgs []cf.Organization, err error)
	FindByName(name string) (org cf.Organization, err error)
}

type CloudControllerOrganizationRepository struct {
	config    *configuration.Configuration
	apiClient ApiClient
}

func NewCloudControllerOrganizationRepository(config *configuration.Configuration, apiClient ApiClient) (repo CloudControllerOrganizationRepository) {
	repo.config = config
	repo.apiClient = apiClient
	return
}

func (repo CloudControllerOrganizationRepository) FindAll() (orgs []cf.Organization, err error) {
	path := repo.config.Target + "/v2/organizations"
	request, err := NewRequest("GET", path, repo.config.AccessToken, nil)
	if err != nil {
		return
	}
	response := new(ApiResponse)

	_, err = repo.apiClient.PerformRequestAndParseResponse(request, response)

	if err != nil {
		return
	}

	for _, r := range response.Resources {
		orgs = append(orgs, cf.Organization{r.Entity.Name, r.Metadata.Guid})
	}

	return
}

func (repo CloudControllerOrganizationRepository) FindByName(name string) (org cf.Organization, err error) {
	orgs, err := repo.FindAll()
	lowerName := strings.ToLower(name)

	if err != nil {
		return
	}

	for _, o := range orgs {
		if strings.ToLower(o.Name) == lowerName {
			return o, nil
		}
	}

	err = errors.New("Organization not found")
	return
}
