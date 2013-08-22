package api

import (
	"cf"
	"cf/configuration"
	"errors"
	"strings"
)

type OrganizationRepository interface {
	FindOrganizations(config *configuration.Configuration) (orgs []cf.Organization, err error)
	FindOrganizationByName(config *configuration.Configuration, name string) (org cf.Organization, err error)
}

type CloudControllerOrganizationRepository struct {
}

func (repo CloudControllerOrganizationRepository) FindOrganizations(config *configuration.Configuration) (orgs []cf.Organization, err error) {
	path := config.Target + "/v2/organizations"
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
		orgs = append(orgs, cf.Organization{r.Entity.Name, r.Metadata.Guid})
	}

	return
}

func (repo CloudControllerOrganizationRepository) FindOrganizationByName(config *configuration.Configuration, name string) (org cf.Organization, err error) {
	orgs, err := repo.FindOrganizations(config)
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
