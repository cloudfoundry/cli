package api

import (
	"cf"
	"cf/configuration"
	"errors"
	"strings"
)

type OrganizationRepository interface {
	FindAll(config *configuration.Configuration) (orgs []cf.Organization, err error)
	FindByName(config *configuration.Configuration, name string) (org cf.Organization, err error)
}

type CloudControllerOrganizationRepository struct {
}

func (repo CloudControllerOrganizationRepository) FindAll(config *configuration.Configuration) (orgs []cf.Organization, err error) {
	path := config.Target + "/v2/organizations"
	request, err := NewRequest("GET", path, config.AccessToken, nil)
	if err != nil {
		return
	}
	response := new(ApiResponse)

	_, err = PerformRequestAndParseResponse(request, response)

	if err != nil {
		return
	}

	for _, r := range response.Resources {
		orgs = append(orgs, cf.Organization{r.Entity.Name, r.Metadata.Guid})
	}

	return
}

func (repo CloudControllerOrganizationRepository) FindByName(config *configuration.Configuration, name string) (org cf.Organization, err error) {
	orgs, err := repo.FindAll(config)
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
