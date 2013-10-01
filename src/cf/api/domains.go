package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"fmt"
	"strings"
)

type DomainRepository interface {
	FindAll() (domains []cf.Domain, apiStatus net.ApiStatus)
	FindByName(name string) (domain cf.Domain, apiStatus net.ApiStatus)
	Create(domainToCreate cf.Domain, owningOrg cf.Organization) (createdDomain cf.Domain, apiStatus net.ApiStatus)
}

type CloudControllerDomainRepository struct {
	config  configuration.Configuration
	gateway net.Gateway
}

func NewCloudControllerDomainRepository(config configuration.Configuration, gateway net.Gateway) (repo CloudControllerDomainRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (repo CloudControllerDomainRepository) FindAll() (domains []cf.Domain, apiStatus net.ApiStatus) {
	path := fmt.Sprintf("%s/v2/spaces/%s/domains", repo.config.Target, repo.config.Space.Guid)
	request, apiStatus := repo.gateway.NewRequest("GET", path, repo.config.AccessToken, nil)
	if apiStatus.IsError() {
		return
	}

	response := new(ApiResponse)
	_, apiStatus = repo.gateway.PerformRequestForJSONResponse(request, response)
	if apiStatus.IsError() {
		return
	}

	for _, r := range response.Resources {
		domains = append(domains, cf.Domain{r.Entity.Name, r.Metadata.Guid})
	}

	return
}

func (repo CloudControllerDomainRepository) FindByName(name string) (domain cf.Domain, apiStatus net.ApiStatus) {
	domains, apiStatus := repo.FindAll()

	if apiStatus.IsError() {
		return
	}

	if name == "" {
		domain = domains[0]
	} else {
		apiStatus = net.NewApiStatusWithMessage("Could not find domain with name %s", name)

		for _, d := range domains {
			if d.Name == strings.ToLower(name) {
				domain = d
				apiStatus = net.ApiStatus{}
			}
		}
	}

	return
}

func (repo CloudControllerDomainRepository) Create(domainToCreate cf.Domain, owningOrg cf.Organization) (createdDomain cf.Domain, apiStatus net.ApiStatus) {
	path := repo.config.Target + "/v2/domains"
	data := fmt.Sprintf(
		`{"name":"%s","wildcard":true,"owning_organization_guid":"%s"}`, domainToCreate.Name, owningOrg.Guid,
	)

	request, apiStatus := repo.gateway.NewRequest("POST", path, repo.config.AccessToken, strings.NewReader(data))
	if apiStatus.IsError() {
		return
	}

	resource := new(Resource)
	_, apiStatus = repo.gateway.PerformRequestForJSONResponse(request, resource)
	if apiStatus.IsError() {
		return
	}

	createdDomain.Guid = resource.Metadata.Guid
	createdDomain.Name = resource.Entity.Name
	return
}
