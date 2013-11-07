package paginator

import (
	"cf/configuration"
	"cf/net"
)

type PaginatedOrganizationResources struct {
	Resources []OrganizationResource
	NextUrl   string `json:"next_url"`
}

type OrganizationResource struct {
	Entity OrganizationEntity
}

type OrganizationEntity struct {
	Name string
}

type OrganizationPaginator struct {
	config  *configuration.Configuration
	gateway net.Gateway
	nextUrl string
}

func NewOrganizationPaginator(config *configuration.Configuration, gateway net.Gateway) (p *OrganizationPaginator) {
	return &OrganizationPaginator{
		config:  config,
		gateway: gateway,
		nextUrl: "/v2/organizations",
	}
}

func (p *OrganizationPaginator) Next() (results []string, apiResponse net.ApiResponse) {
	orgResources := new(PaginatedOrganizationResources)
	path := p.config.Target + p.nextUrl

	apiResponse = p.gateway.GetResource(path, p.config.AccessToken, orgResources)
	if apiResponse.IsNotSuccessful() {
		return
	}

	for _, r := range orgResources.Resources {
		results = append(results, r.Entity.Name)
	}

	p.nextUrl = orgResources.NextUrl
	return
}

func (p *OrganizationPaginator) HasNext() bool {
	return p.nextUrl != ""
}
