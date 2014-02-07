package api

import (
	"cf"
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

type PaginatedOrganizationResources struct {
	Resources []OrganizationResource
	NextUrl   string `json:"next_url"`
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
		domains = append(domains, d.ToFields().(models.DomainFields))
	}
	org.Domains = domains

	return
}

type OrganizationRepository interface {
	ListOrgs(stop chan bool) (orgsChan chan []models.Organization, statusChan chan net.ApiResponse)
	FindByName(name string) (org models.Organization, apiResponse net.ApiResponse)
	Create(name string) (apiResponse net.ApiResponse)
	Rename(orgGuid string, name string) (apiResponse net.ApiResponse)
	Delete(orgGuid string) (apiResponse net.ApiResponse)
}

type CloudControllerOrganizationRepository struct {
	config  *configuration.Configuration
	gateway net.Gateway
}

func NewCloudControllerOrganizationRepository(config *configuration.Configuration, gateway net.Gateway) (repo CloudControllerOrganizationRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (repo CloudControllerOrganizationRepository) ListOrgs(stop chan bool) (orgsChan chan []models.Organization, statusChan chan net.ApiResponse) {
	orgsChan = make(chan []models.Organization, 4)
	statusChan = make(chan net.ApiResponse, 1)

	go func() {
		path := "/v2/organizations"

	loop:
		for path != "" {
			select {
			case <-stop:
				break loop
			default:
				var (
					organizations []models.Organization
					apiResponse   net.ApiResponse
				)
				organizations, path, apiResponse = repo.findNextWithPath(path)
				if apiResponse.IsNotSuccessful() {
					statusChan <- apiResponse
					close(orgsChan)
					close(statusChan)
					return
				}

				if len(organizations) > 0 {
					orgsChan <- organizations
				}
			}
		}
		close(orgsChan)
		close(statusChan)
		cf.WaitForClose(stop)
	}()

	return
}

func (repo CloudControllerOrganizationRepository) findNextWithPath(path string) (orgs []models.Organization, nextUrl string, apiResponse net.ApiResponse) {
	orgResources := new(PaginatedOrganizationResources)

	apiResponse = repo.gateway.GetResource(repo.config.Target+path, repo.config.AccessToken, orgResources)
	if apiResponse.IsNotSuccessful() {
		return
	}

	nextUrl = orgResources.NextUrl

	for _, r := range orgResources.Resources {
		orgs = append(orgs, r.ToModel())
	}
	return
}

func (repo CloudControllerOrganizationRepository) FindByName(name string) (org models.Organization, apiResponse net.ApiResponse) {
	path := fmt.Sprintf("/v2/organizations?q=%s&inline-relations-depth=1", url.QueryEscape("name:"+strings.ToLower(name)))

	orgs, _, apiResponse := repo.findNextWithPath(path)
	if apiResponse.IsNotSuccessful() {
		return
	}

	if len(orgs) != 1 {
		apiResponse = net.NewNotFoundApiResponse("Org %s not found", name)
		return
	}

	org = orgs[0]
	return
}

func (repo CloudControllerOrganizationRepository) Create(name string) (apiResponse net.ApiResponse) {
	url := repo.config.Target + "/v2/organizations"
	data := fmt.Sprintf(`{"name":"%s"}`, name)
	return repo.gateway.CreateResource(url, repo.config.AccessToken, strings.NewReader(data))
}

func (repo CloudControllerOrganizationRepository) Rename(orgGuid string, name string) (apiResponse net.ApiResponse) {
	url := fmt.Sprintf("%s/v2/organizations/%s", repo.config.Target, orgGuid)
	data := fmt.Sprintf(`{"name":"%s"}`, name)
	return repo.gateway.UpdateResource(url, repo.config.AccessToken, strings.NewReader(data))
}

func (repo CloudControllerOrganizationRepository) Delete(orgGuid string) (apiResponse net.ApiResponse) {
	url := fmt.Sprintf("%s/v2/organizations/%s?recursive=true", repo.config.Target, orgGuid)
	return repo.gateway.DeleteResource(url, repo.config.AccessToken)
}
