package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"fmt"
	"strings"
)

type DomainRepository interface {
	FindAllInCurrentSpace() (domains []cf.Domain, apiResponse net.ApiResponse)
	FindAllByOrg(org cf.Organization) (domains []cf.Domain, apiResponse net.ApiResponse)
	FindByNameInCurrentSpace(name string) (domain cf.Domain, apiResponse net.ApiResponse)
	FindByNameInOrg(name string, owningOrg cf.Organization) (domain cf.Domain, apiResponse net.ApiResponse)
	Create(domainToCreate cf.Domain, owningOrg cf.Organization) (createdDomain cf.Domain, apiResponse net.ApiResponse)
	Share(domainToShare cf.Domain) (apiResponse net.ApiResponse)
	MapDomain(domain cf.Domain, space cf.Space) (apiResponse net.ApiResponse)
	UnmapDomain(domain cf.Domain, space cf.Space) (apiResponse net.ApiResponse)
	DeleteDomain(domain cf.Domain) (apiResponse net.ApiResponse)
}

type CloudControllerDomainRepository struct {
	config  *configuration.Configuration
	gateway net.Gateway
}

func NewCloudControllerDomainRepository(config *configuration.Configuration, gateway net.Gateway) (repo CloudControllerDomainRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (repo CloudControllerDomainRepository) FindAllInCurrentSpace() (domains []cf.Domain, apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/spaces/%s/domains", repo.config.Target, repo.config.Space.Guid)
	request, apiResponse := repo.gateway.NewRequest("GET", path, repo.config.AccessToken, nil)
	if apiResponse.IsNotSuccessful() {
		return
	}

	response := new(ApiResponse)
	_, apiResponse = repo.gateway.PerformRequestForJSONResponse(request, response)
	if apiResponse.IsNotSuccessful() {
		return
	}

	for _, r := range response.Resources {
		domains = append(domains, cf.Domain{Name: r.Entity.Name, Guid: r.Metadata.Guid})
	}

	return
}

func (repo CloudControllerDomainRepository) FindAllByOrg(org cf.Organization) (domains []cf.Domain, apiResponse net.ApiResponse) {
	orgGuid := org.Guid

	path := fmt.Sprintf("%s/v2/organizations/%s/domains?inline-relations-depth=1", repo.config.Target, orgGuid)
	request, apiResponse := repo.gateway.NewRequest("GET", path, repo.config.AccessToken, nil)
	if apiResponse.IsNotSuccessful() {
		return
	}

	response := new(DomainApiResponse)
	_, apiResponse = repo.gateway.PerformRequestForJSONResponse(request, response)
	if apiResponse.IsNotSuccessful() {
		return
	}

	for _, r := range response.Resources {
		domain := cf.Domain{
			Name: r.Entity.Name,
			Guid: r.Metadata.Guid,
		}
		domain.Shared = r.Entity.OwningOrganizationGuid == ""

		for _, space := range r.Entity.Spaces {
			domain.Spaces = append(domain.Spaces, cf.Space{
				Name: space.Entity.Name,
				Guid: space.Metadata.Guid,
			})
		}
		domains = append(domains, domain)
	}

	return
}

func (repo CloudControllerDomainRepository) FindByNameInCurrentSpace(name string) (domain cf.Domain, apiResponse net.ApiResponse) {
	domains, apiResponse := repo.FindAllInCurrentSpace()

	if apiResponse.IsNotSuccessful() {
		return
	}

	domainIndex := indexOfDomain(domains, name)

	if domainIndex >= 0 {
		domain = domains[domainIndex]
	} else {
		apiResponse = net.NewNotFoundApiResponse("%s %s not found", "Domain", name)
	}

	return
}

func (repo CloudControllerDomainRepository) Create(domainToCreate cf.Domain, owningOrg cf.Organization) (createdDomain cf.Domain, apiResponse net.ApiResponse) {
	path := repo.config.Target + "/v2/domains"
	data := fmt.Sprintf(
		`{"name":"%s","wildcard":true,"owning_organization_guid":"%s"}`, domainToCreate.Name, owningOrg.Guid,
	)

	request, apiResponse := repo.gateway.NewRequest("POST", path, repo.config.AccessToken, strings.NewReader(data))
	if apiResponse.IsNotSuccessful() {
		return
	}

	resource := new(Resource)
	_, apiResponse = repo.gateway.PerformRequestForJSONResponse(request, resource)
	if apiResponse.IsNotSuccessful() {
		return
	}

	createdDomain.Guid = resource.Metadata.Guid
	createdDomain.Name = resource.Entity.Name
	return
}

func (repo CloudControllerDomainRepository) Share(domainToShare cf.Domain) (apiResponse net.ApiResponse) {
	path := repo.config.Target + "/v2/domains"
	data := fmt.Sprintf(`{"name":"%s","wildcard":true,"shared":true}`, domainToShare.Name)

	request, apiResponse := repo.gateway.NewRequest("POST", path, repo.config.AccessToken, strings.NewReader(data))
	if apiResponse.IsNotSuccessful() {
		return
	}

	apiResponse = repo.gateway.PerformRequest(request)

	return
}

func (repo CloudControllerDomainRepository) MapDomain(domain cf.Domain, space cf.Space) (apiResponse net.ApiResponse) {
	return repo.changeDomain("PUT", domain, space)
}

func (repo CloudControllerDomainRepository) UnmapDomain(domain cf.Domain, space cf.Space) (apiResponse net.ApiResponse) {
	return repo.changeDomain("DELETE", domain, space)
}

func (repo CloudControllerDomainRepository) changeDomain(verb string, domain cf.Domain, space cf.Space) (apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/spaces/%s/domains/%s", repo.config.Target, space.Guid, domain.Guid)

	request, apiResponse := repo.gateway.NewRequest(verb, path, repo.config.AccessToken, nil)
	if apiResponse.IsNotSuccessful() {
		return
	}

	apiResponse = repo.gateway.PerformRequest(request)
	return
}

func (repo CloudControllerDomainRepository) FindByNameInOrg(name string, owningOrg cf.Organization) (domain cf.Domain, apiResponse net.ApiResponse) {
	domains, apiResponse := repo.FindAllByOrg(owningOrg)
	if apiResponse.IsNotSuccessful() {
		return
	}

	domainIndex := indexOfDomain(domains, name)
	if domainIndex >= 0 {
		domain = domains[domainIndex]
	} else {
		apiResponse = net.NewNotFoundApiResponse("%s %s not found", "Domain", name)
	}

	return
}

func (repo CloudControllerDomainRepository) DeleteDomain(domain cf.Domain) (apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/domains/%s?recursive=true", repo.config.Target, domain.Guid)
	request, apiResponse := repo.gateway.NewRequest("DELETE", path, repo.config.AccessToken, nil)
	if apiResponse.IsNotSuccessful() {
		return
	}

	apiResponse = repo.gateway.PerformRequest(request)
	return
}
func indexOfDomain(domains []cf.Domain, domainName string) int {
	domainName = strings.ToLower(domainName)

	if len(domains) > 0 && domainName == "" {
		return 0
	}

	for i, d := range domains {
		if strings.ToLower(d.Name) == domainName {
			return i
		}
	}

	return -1
}
