package strategy

type DomainsEndpointStrategy interface {
	OrgDomainURL(orgGuid, name string) string
	DomainURL(name string) string
	OrgDomainsURL(orgGuid string) string
	PrivateDomainsURL() string
	SharedDomainsURL() string
	DeleteDomainURL(guid string) string
	DeleteSharedDomainURL(guid string) string
}

type domainsEndpointStrategy struct{}

func (s domainsEndpointStrategy) DomainURL(name string) string {
	return buildURL(v2("domains"), params{
		inlineRelationsDepth: 1,
		q:                    map[string]string{"name": name},
	})
}

func (s domainsEndpointStrategy) OrgDomainsURL(orgGuid string) string {
	return v2("organizations", orgGuid, "domains")
}

func (s domainsEndpointStrategy) OrgDomainURL(orgGuid, name string) string {
	return buildURL(s.OrgDomainsURL(orgGuid), params{
		inlineRelationsDepth: 1,
		q:                    map[string]string{"name": name},
	})
}

func (s domainsEndpointStrategy) PrivateDomainsURL() string {
	return v2("domains")
}

func (s domainsEndpointStrategy) SharedDomainsURL() string {
	return v2("domains")
}

func (s domainsEndpointStrategy) DeleteDomainURL(guid string) string {
	return buildURL(v2("domains", guid), params{recursive: true})
}

func (s domainsEndpointStrategy) DeleteSharedDomainURL(guid string) string {
	return buildURL(v2("domains", guid), params{recursive: true})
}

type separatedDomainsEndpointStrategy struct {
	domainsEndpointStrategy
}

func (s separatedDomainsEndpointStrategy) PrivateDomainsURL() string {
	return v2("private_domains")
}

func (s separatedDomainsEndpointStrategy) SharedDomainsURL() string {
	return v2("shared_domains")
}

func (s separatedDomainsEndpointStrategy) DeleteDomainURL(guid string) string {
	return buildURL(v2("private_domains", guid), params{recursive: true})
}

func (s separatedDomainsEndpointStrategy) DeleteSharedDomainURL(guid string) string {
	return buildURL(v2("shared_domains", guid), params{recursive: true})
}
