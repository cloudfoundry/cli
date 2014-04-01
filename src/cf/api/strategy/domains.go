package strategy

import "fmt"

type DomainsEndpointStrategy interface {
	DomainsURL(orgGuid string) string
}

type globalDomainsEndpointStrategy struct{}
type orgScopedDomainsEndpointStrategy struct{}

func (_ globalDomainsEndpointStrategy) DomainsURL(orgGuid string) string {
	return "/v2/domains"
}

func (_ orgScopedDomainsEndpointStrategy) DomainsURL(orgGuid string) string {
	return fmt.Sprintf("/v2/organizations/%s/domains", orgGuid)
}
