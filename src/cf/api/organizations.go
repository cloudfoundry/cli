package api

import (
	"cf"
	"cf/configuration"
	"fmt"
	"strings"
)

type OrganizationRepository interface {
	FindAll() (orgs []cf.Organization, apiErr *ApiError)
	FindByName(name string) (org cf.Organization, apiErr *ApiError)
	Create(name string) (apiErr *ApiError)
	Rename(org cf.Organization, name string) (apiErr *ApiError)
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

func (repo CloudControllerOrganizationRepository) FindAll() (orgs []cf.Organization, apiErr *ApiError) {
	path := repo.config.Target + "/v2/organizations"
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
		orgs = append(orgs, cf.Organization{r.Entity.Name, r.Metadata.Guid})
	}

	return
}

func (repo CloudControllerOrganizationRepository) FindByName(name string) (org cf.Organization, apiErr *ApiError) {
	orgs, apiErr := repo.FindAll()
	lowerName := strings.ToLower(name)

	if apiErr != nil {
		return
	}

	for _, o := range orgs {
		if strings.ToLower(o.Name) == lowerName {
			return o, nil
		}
	}

	apiErr = NewApiErrorWithMessage("Organization not found")
	return
}

func (repo CloudControllerOrganizationRepository) Create(name string) (apiErr *ApiError) {
	path := repo.config.Target + "/v2/organizations"
	data := fmt.Sprintf(
		`{"name":"%s"}`, name,
	)
	request, apiErr := NewRequest("POST", path, repo.config.AccessToken, strings.NewReader(data))
	if apiErr != nil {
		return
	}

	apiErr = repo.apiClient.PerformRequest(request)
	return
}

func (repo CloudControllerOrganizationRepository) Rename(org cf.Organization, name string) (apiErr *ApiError) {
	path := fmt.Sprintf("%s/v2/organizations/%s", repo.config.Target, org.Guid)
	data := fmt.Sprintf(`{"name":"%s"}`, name)
	request, apiErr := NewRequest("PUT", path, repo.config.AccessToken, strings.NewReader(data))
	if apiErr != nil {
		return
	}

	apiErr = repo.apiClient.PerformRequest(request)
	return
}
