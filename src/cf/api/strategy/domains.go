package strategy

type DomainsEndpointStrategy interface {
	DomainsURL(orgGuid string) string
	CreatePrivateDomainURL() string
}

type orgScopedDomainsEndpointStrategy struct{}

func (_ orgScopedDomainsEndpointStrategy) DomainsURL(orgGuid string) string {
	return "/v2/organizations/"+orgGuid+"/domains"
}

func (_ orgScopedDomainsEndpointStrategy) CreatePrivateDomainURL() string {
	return "/v2/private_domains"
}

type globalDomainsEndpointStrategy struct{}

func (_ globalDomainsEndpointStrategy) DomainsURL(orgGuid string) string {
	return "/v2/domains"
}

func (_ globalDomainsEndpointStrategy) CreatePrivateDomainURL() string {
	return "/v2/domains"
}
