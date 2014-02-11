package api

import (
	"cf/configuration"
	"cf/models"
	"cf/net"
	"fmt"
	"net/url"
	"strings"
)

type OrganizationEntity struct {
	Name            string
	QuotaDefinition QuotaResource `json:"quota_definition"`
	Spaces          []SpaceResource
	Domains         []DomainResource
}

type OrganizationResource struct {
	Resource
	Entity OrganizationEntity
}

func (resource OrganizationResource) ToFields() (fields models.OrganizationFields) {
	fields.Name = resource.Entity.Name
	fields.Guid = resource.Metadata.Guid

	fields.QuotaDefinition = resource.Entity.QuotaDefinition.ToFields()
	return
}

func (resource OrganizationResource) ToModel() (org models.Organization) {
	org.OrganizationFields = resource.ToFields()

	spaces := []models.SpaceFields{}
	for _, s := range resource.Entity.Spaces {
		spaces = append(spaces, s.ToFields())
	}
	org.Spaces = spaces

	domains := []models.DomainFields{}
	for _, d := range resource.Entity.Domains {
		domains = append(domains, d.ToFields())
	}
	org.Domains = domains

	return
}

type OrganizationRepository interface {
	ListOrgs(func(models.Organization) bool) (apiResponse net.ApiResponse)
	FindByName(name string) (org models.Organization, apiResponse net.ApiResponse)
	Create(name string) (apiResponse net.ApiResponse)
	Rename(orgGuid string, name string) (apiResponse net.ApiResponse)
	Delete(orgGuid string) (apiResponse net.ApiResponse)
}

type CloudControllerOrganizationRepository struct {
	config  configuration.Reader
	gateway net.Gateway
}

func NewCloudControllerOrganizationRepository(config configuration.Reader, gateway net.Gateway) (repo CloudControllerOrganizationRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (repo CloudControllerOrganizationRepository) ListOrgs(cb func(models.Organization) bool) (apiResponse net.ApiResponse) {
	return repo.gateway.ListPaginatedResources(
		repo.config.ApiEndpoint(),
		repo.config.AccessToken(),
		"/v2/organizations",
		OrganizationResource{},
		func(resource interface{}) bool {
			return cb(resource.(OrganizationResource).ToModel())
		})
}

func (repo CloudControllerOrganizationRepository) FindByName(name string) (org models.Organization, apiResponse net.ApiResponse) {
	found := false
	apiResponse = repo.gateway.ListPaginatedResources(
		repo.config.ApiEndpoint(),
		repo.config.AccessToken(),
		fmt.Sprintf("/v2/organizations?q=%s&inline-relations-depth=1", url.QueryEscape("name:"+strings.ToLower(name))),
		OrganizationResource{},
		func(resource interface{}) bool {
			org = resource.(OrganizationResource).ToModel()
			found = true
			return false
		})

	if !found {
		apiResponse = net.NewNotFoundApiResponse("Organization %s not found", name)
	}

	return
}

func (repo CloudControllerOrganizationRepository) Create(name string) (apiResponse net.ApiResponse) {
	url := repo.config.ApiEndpoint() + "/v2/organizations"
	data := fmt.Sprintf(`{"name":"%s"}`, name)
	return repo.gateway.CreateResource(url, repo.config.AccessToken(), strings.NewReader(data))
}

func (repo CloudControllerOrganizationRepository) Rename(orgGuid string, name string) (apiResponse net.ApiResponse) {
	url := fmt.Sprintf("%s/v2/organizations/%s", repo.config.ApiEndpoint(), orgGuid)
	data := fmt.Sprintf(`{"name":"%s"}`, name)
	return repo.gateway.UpdateResource(url, repo.config.AccessToken(), strings.NewReader(data))
}

func (repo CloudControllerOrganizationRepository) Delete(orgGuid string) (apiResponse net.ApiResponse) {
	url := fmt.Sprintf("%s/v2/organizations/%s?recursive=true", repo.config.ApiEndpoint(), orgGuid)
	return repo.gateway.DeleteResource(url, repo.config.AccessToken())
}
