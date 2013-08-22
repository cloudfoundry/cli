package api

import (
	"cf"
	"cf/configuration"
	"errors"
	"net/http"
	"strings"
)

type OrganizationRepository interface {
	FindOrganizations(config *configuration.Configuration) (orgs []cf.Organization, err error)
	FindOrganizationByName(config *configuration.Configuration, name string) (org cf.Organization, err error)
}

type CloudControllerOrganizationRepository struct {
}

func (repo CloudControllerOrganizationRepository) FindOrganizations(config *configuration.Configuration) (orgs []cf.Organization, err error) {
	request, err := http.NewRequest("GET", config.Target+"/v2/organizations", nil)
	if err != nil {
		return
	}
	request.Header.Set("Authorization", config.AccessToken)

	type Metadata struct {
		Guid string
	}

	type Entity struct {
		Name string
	}

	type Resource struct {
		Metadata Metadata
		Entity   Entity
	}

	type OrganizationsResponse struct {
		Resources []Resource
	}

	response := new(OrganizationsResponse)

	err = PerformRequest(request, response)

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
