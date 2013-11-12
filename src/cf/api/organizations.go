package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"fmt"
	"strings"
)

type PaginatedOrganizationResources struct {
	Resources []OrganizationResource
	NextUrl   string `json:"next_url"`
}

type OrganizationResource struct {
	Resource
	Entity OrganizationEntity
}

type OrganizationEntity struct {
	Name    string
	Spaces  []Resource
	Domains []Resource
}

type OrganizationRepository interface {
	ListOrgs(stop chan bool) (orgsChan chan []cf.Organization, statusChan chan net.ApiResponse)
	FindByName(name string) (org cf.Organization, apiResponse net.ApiResponse)
	Create(name string) (apiResponse net.ApiResponse)
	Rename(org cf.Organization, name string) (apiResponse net.ApiResponse)
	Delete(org cf.Organization) (apiResponse net.ApiResponse)
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

func (repo CloudControllerOrganizationRepository) ListOrgs(stop chan bool) (orgsChan chan []cf.Organization, statusChan chan net.ApiResponse) {
	orgsChan = make(chan []cf.Organization, 4)
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
					organizations []cf.Organization
					apiResponse   net.ApiResponse
				)
				organizations, path, apiResponse = repo.findNextWithPath(path)
				if apiResponse.IsNotSuccessful() {
					statusChan <- apiResponse
					close(orgsChan)
					close(statusChan)
					return
				}

				orgsChan <- organizations
			}
		}
		close(orgsChan)
		close(statusChan)
		cf.WaitForClose(stop)
	}()

	return
}

func (repo CloudControllerOrganizationRepository) findNextWithPath(path string) (orgs []cf.Organization, nextUrl string, apiResponse net.ApiResponse) {
	orgResources := new(PaginatedOrganizationResources)

	apiResponse = repo.gateway.GetResource(repo.config.Target+path, repo.config.AccessToken, orgResources)
	if apiResponse.IsNotSuccessful() {
		return
	}

	nextUrl = orgResources.NextUrl

	for _, r := range orgResources.Resources {
		spaces := []cf.Space{}
		for _, s := range r.Entity.Spaces {
			spaces = append(spaces, cf.Space{Name: s.Entity.Name, Guid: s.Metadata.Guid})
		}

		domains := []cf.Domain{}
		for _, d := range r.Entity.Domains {
			domains = append(domains, cf.Domain{Name: d.Entity.Name, Guid: d.Metadata.Guid})
		}

		orgs = append(orgs, cf.Organization{
			Name:    r.Entity.Name,
			Guid:    r.Metadata.Guid,
			Spaces:  spaces,
			Domains: domains,
		})
	}
	return
}

func (repo CloudControllerOrganizationRepository) FindByName(name string) (org cf.Organization, apiResponse net.ApiResponse) {
	path := fmt.Sprintf("/v2/organizations?q=name%s&inline-relations-depth=1", "%3A"+strings.ToLower(name))

	orgs, _, apiResponse := repo.findNextWithPath(path)
	if apiResponse.IsNotSuccessful() {
		return
	}

	if len(orgs) == 0 {
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

func (repo CloudControllerOrganizationRepository) Rename(org cf.Organization, name string) (apiResponse net.ApiResponse) {
	url := fmt.Sprintf("%s/v2/organizations/%s", repo.config.Target, org.Guid)
	data := fmt.Sprintf(`{"name":"%s"}`, name)
	return repo.gateway.UpdateResource(url, repo.config.AccessToken, strings.NewReader(data))
}

func (repo CloudControllerOrganizationRepository) Delete(org cf.Organization) (apiResponse net.ApiResponse) {
	url := fmt.Sprintf("%s/v2/organizations/%s?recursive=true", repo.config.Target, org.Guid)
	return repo.gateway.DeleteResource(url, repo.config.AccessToken)
}
