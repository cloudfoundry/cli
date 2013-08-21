package api

import (
	"cf/configuration"
	"net/http"
)

type Organization struct {
	Name string
	Guid string
}

type OrganizationRepository interface {
	FindOrganizations(config *configuration.Configuration) (orgs []Organization, err error)
}

type CloudControllerOrganizationRepository struct {
}

func (repo CloudControllerOrganizationRepository) FindOrganizations(config *configuration.Configuration) (orgs []Organization, err error) {
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
		orgs = append(orgs, Organization{r.Entity.Name, r.Metadata.Guid})
	}

	return
}
