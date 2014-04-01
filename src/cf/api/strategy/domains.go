package strategy

type DomainsEndpointStrategy interface {
	DomainsURL(orgGuid string) string
	PrivateDomainsURL() string
	SharedDomainsURL() string
	DeleteDomainURL(guid string) string
	DeleteSharedDomainURL(guid string) string
}

type orgScopedDomainsEndpointStrategy struct{}

func (s orgScopedDomainsEndpointStrategy) DomainsURL(orgGuid string) string {
	return "/v2/organizations/" + orgGuid + "/domains"
}

func (s orgScopedDomainsEndpointStrategy) PrivateDomainsURL() string {
	return "/v2/private_domains"
}

func (s orgScopedDomainsEndpointStrategy) SharedDomainsURL() string {
	return "/v2/shared_domains"
}

func (s orgScopedDomainsEndpointStrategy) DeleteDomainURL(guid string) string {
	return buildURL(s.PrivateDomainsURL()+"/"+guid, query{recursive: true})
}

func (s orgScopedDomainsEndpointStrategy) DeleteSharedDomainURL(guid string) string {
	return buildURL(s.SharedDomainsURL()+"/"+guid, query{recursive: true})
}

type globalDomainsEndpointStrategy struct{}

func (s globalDomainsEndpointStrategy) DomainsURL(orgGuid string) string {
	return "/v2/domains"
}

func (s globalDomainsEndpointStrategy) PrivateDomainsURL() string {
	return "/v2/domains"
}

func (s globalDomainsEndpointStrategy) SharedDomainsURL() string {
	return "/v2/domains"
}

func (s globalDomainsEndpointStrategy) DeleteDomainURL(guid string) string {
	return buildURL(s.PrivateDomainsURL()+"/"+guid, query{recursive: true})
}

func (s globalDomainsEndpointStrategy) DeleteSharedDomainURL(guid string) string {
	return buildURL(s.SharedDomainsURL()+"/"+guid, query{recursive: true})
}
