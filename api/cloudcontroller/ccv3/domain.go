package ccv3

import (
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/resources"
)

type SharedOrgs struct {
	GUIDs []string
}

func (sharedOrgs SharedOrgs) MarshalJSON() ([]byte, error) {
	type Org struct {
		GUID string `json:"guid,omitempty"`
	}

	type Data = []Org

	type sharedOrgsRelationship struct {
		Data Data `json:"data"`
	}

	var orgs []Org
	for _, sharedOrgGUID := range sharedOrgs.GUIDs {
		orgs = append(orgs, Org{GUID: sharedOrgGUID})
	}

	relationship := sharedOrgsRelationship{
		Data: orgs,
	}

	return json.Marshal(relationship)
}

func (sharedOrgs *SharedOrgs) UnmarshalJSON(data []byte) error {
	var alias struct {
		Data []struct {
			GUID string `json:"guid,omitempty"`
		} `json:"data,omitempty"`
	}

	err := cloudcontroller.DecodeJSON(data, &alias)
	if err != nil {
		return err
	}

	var guids []string
	for _, org := range alias.Data {
		guids = append(guids, org.GUID)
	}

	sharedOrgs.GUIDs = guids
	return nil
}

// CheckRoute checks whether the route with the given domain GUID, hostname,
// and path exists in the foundation.
func (client Client) CheckRoute(domainGUID string, hostname string, path string) (bool, Warnings, error) {
	var query []Query

	if hostname != "" {
		query = append(query, Query{Key: HostFilter, Values: []string{hostname}})
	}

	if path != "" {
		query = append(query, Query{Key: PathFilter, Values: []string{path}})
	}

	var responseBody struct {
		MatchingRoute bool `json:"matching_route"`
	}

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.GetDomainRouteReservationsRequest,
		URIParams:    internal.Params{"domain_guid": domainGUID},
		Query:        query,
		ResponseBody: &responseBody,
	})

	return responseBody.MatchingRoute, warnings, err
}

func (client Client) CreateDomain(domain resources.Domain) (resources.Domain, Warnings, error) {
	var responseBody resources.Domain

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.PostDomainRequest,
		RequestBody:  domain,
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

func (client Client) DeleteDomain(domainGUID string) (JobURL, Warnings, error) {
	jobURL, warnings, err := client.MakeRequest(RequestParams{
		RequestName: internal.DeleteDomainRequest,
		URIParams:   internal.Params{"domain_guid": domainGUID},
	})

	return jobURL, warnings, err
}

// GetDomain returns a domain with the given GUID.
func (client *Client) GetDomain(domainGUID string) (resources.Domain, Warnings, error) {
	var responseBody resources.Domain

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.GetDomainRequest,
		URIParams:    internal.Params{"domain_guid": domainGUID},
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

func (client Client) GetDomains(query ...Query) ([]resources.Domain, Warnings, error) {
	var domains []resources.Domain

	_, warnings, err := client.MakeListRequest(RequestParams{
		RequestName:  internal.GetDomainsRequest,
		Query:        query,
		ResponseBody: resources.Domain{},
		AppendToList: func(item interface{}) error {
			domains = append(domains, item.(resources.Domain))
			return nil
		},
	})

	return domains, warnings, err
}

func (client Client) GetOrganizationDomains(orgGUID string, query ...Query) ([]resources.Domain, Warnings, error) {
	var domains []resources.Domain

	_, warnings, err := client.MakeListRequest(RequestParams{
		URIParams:    internal.Params{"organization_guid": orgGUID},
		RequestName:  internal.GetOrganizationDomainsRequest,
		Query:        query,
		ResponseBody: resources.Domain{},
		AppendToList: func(item interface{}) error {
			domains = append(domains, item.(resources.Domain))
			return nil
		},
	})

	return domains, warnings, err
}

func (client Client) SharePrivateDomainToOrgs(domainGuid string, sharedOrgs SharedOrgs) (Warnings, error) {
	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName: internal.SharePrivateDomainRequest,
		URIParams:   internal.Params{"domain_guid": domainGuid},
		RequestBody: sharedOrgs,
	})

	return warnings, err
}

func (client Client) UnsharePrivateDomainFromOrg(domainGuid string, orgGUID string) (Warnings, error) {
	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName: internal.DeleteSharedOrgFromDomainRequest,
		URIParams:   internal.Params{"domain_guid": domainGuid, "org_guid": orgGUID},
	})

	return warnings, err
}
