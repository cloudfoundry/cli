package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"fmt"
	"strings"
)

type DomainRepository interface {
	FindAllInCurrentSpace() (domains []cf.Domain, apiStatus net.ApiStatus)
	FindAllByOrg(org cf.Organization) (domains []cf.Domain, apiStatus net.ApiStatus)
	FindByNameInCurrentSpace(name string) (domain cf.Domain, apiStatus net.ApiStatus)
	FindByNameInOrg(name string, owningOrg cf.Organization) (domain cf.Domain, apiStatus net.ApiStatus)
	Create(domainToCreate cf.Domain, owningOrg cf.Organization) (createdDomain cf.Domain, apiStatus net.ApiStatus)
	MapDomain(domain cf.Domain, space cf.Space) (apiStatus net.ApiStatus)
	UnmapDomain(domain cf.Domain, space cf.Space) (apiStatus net.ApiStatus)
	DeleteDomain(domain cf.Domain) (apiStatus net.ApiStatus)
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

func (repo CloudControllerDomainRepository) FindAllInCurrentSpace() (domains []cf.Domain, apiStatus net.ApiStatus) {
	path := fmt.Sprintf("%s/v2/spaces/%s/domains", repo.config.Target, repo.config.Space.Guid)
	request, apiStatus := repo.gateway.NewRequest("GET", path, repo.config.AccessToken, nil)
	if apiStatus.IsNotSuccessful() {
		return
	}

	response := new(ApiResponse)
	_, apiStatus = repo.gateway.PerformRequestForJSONResponse(request, response)
	if apiStatus.IsNotSuccessful() {
		return
	}

	for _, r := range response.Resources {
		domains = append(domains, cf.Domain{Name: r.Entity.Name, Guid: r.Metadata.Guid})
	}

	return
}

func (repo CloudControllerDomainRepository) FindAllByOrg(org cf.Organization) (domains []cf.Domain, apiStatus net.ApiStatus) {
	orgGuid := org.Guid

	path := fmt.Sprintf("%s/v2/organizations/%s/domains?inline-relations-depth=1", repo.config.Target, orgGuid)
	request, apiStatus := repo.gateway.NewRequest("GET", path, repo.config.AccessToken, nil)
	if apiStatus.IsNotSuccessful() {
		return
	}

	response := new(DomainApiResponse)
	_, apiStatus = repo.gateway.PerformRequestForJSONResponse(request, response)
	if apiStatus.IsNotSuccessful() {
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

func (repo CloudControllerDomainRepository) FindByNameInCurrentSpace(name string) (domain cf.Domain, apiStatus net.ApiStatus) {
	domains, apiStatus := repo.FindAllInCurrentSpace()

	if apiStatus.IsNotSuccessful() {
		return
	}

	domainIndex := indexOfDomain(domains, name)

	if domainIndex >= 0 {
		domain = domains[domainIndex]
	} else {
		apiStatus = net.NewNotFoundApiStatus("Domain", name)
	}

	return
}

func (repo CloudControllerDomainRepository) Create(domainToCreate cf.Domain, owningOrg cf.Organization) (createdDomain cf.Domain, apiStatus net.ApiStatus) {
	path := repo.config.Target + "/v2/domains"
	data := fmt.Sprintf(
		`{"name":"%s","wildcard":true,"owning_organization_guid":"%s"}`, domainToCreate.Name, owningOrg.Guid,
	)

	request, apiStatus := repo.gateway.NewRequest("POST", path, repo.config.AccessToken, strings.NewReader(data))
	if apiStatus.IsNotSuccessful() {
		return
	}

	resource := new(Resource)
	_, apiStatus = repo.gateway.PerformRequestForJSONResponse(request, resource)
	if apiStatus.IsNotSuccessful() {
		return
	}

	createdDomain.Guid = resource.Metadata.Guid
	createdDomain.Name = resource.Entity.Name
	return
}

func (repo CloudControllerDomainRepository) MapDomain(domain cf.Domain, space cf.Space) (apiStatus net.ApiStatus) {
	return repo.changeDomain("PUT", domain, space)
}

func (repo CloudControllerDomainRepository) UnmapDomain(domain cf.Domain, space cf.Space) (apiStatus net.ApiStatus) {
	return repo.changeDomain("DELETE", domain, space)
}

func (repo CloudControllerDomainRepository) changeDomain(verb string, domain cf.Domain, space cf.Space) (apiStatus net.ApiStatus) {
	path := fmt.Sprintf("%s/v2/spaces/%s/domains/%s", repo.config.Target, space.Guid, domain.Guid)

	request, apiStatus := repo.gateway.NewRequest(verb, path, repo.config.AccessToken, nil)
	if apiStatus.IsNotSuccessful() {
		return
	}

	apiStatus = repo.gateway.PerformRequest(request)
	return
}

func (repo CloudControllerDomainRepository) FindByNameInOrg(name string, owningOrg cf.Organization) (domain cf.Domain, apiStatus net.ApiStatus) {
	domains, apiStatus := repo.FindAllByOrg(owningOrg)
	if apiStatus.IsNotSuccessful() {
		return
	}

	domainIndex := indexOfDomain(domains, name)
	if domainIndex >= 0 {
		domain = domains[domainIndex]
	} else {
		apiStatus = net.NewNotFoundApiStatus("Domain", name)
	}

	return
}

func (repo CloudControllerDomainRepository) DeleteDomain(domain cf.Domain) (apiStatus net.ApiStatus) {
	path := fmt.Sprintf("%s/v2/domains/%s?recursive=true", repo.config.Target, domain.Guid)
	request, apiStatus := repo.gateway.NewRequest("DELETE", path, repo.config.AccessToken, nil)
	if apiStatus.IsNotSuccessful() {
		return
	}

	apiStatus = repo.gateway.PerformRequest(request)
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
