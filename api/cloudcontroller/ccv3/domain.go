package ccv3

import (
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/types"
)

type Domain struct {
	GUID             string         `json:"guid,omitempty"`
	Name             string         `json:"name"`
	Internal         types.NullBool `json:"internal,omitempty"`
	OrganizationGUID string         `json:"orgguid,omitempty"`

	// Metadata is used for custom tagging of API resources
	Metadata *Metadata `json:"metadata,omitempty"`
}

func (d Domain) MarshalJSON() ([]byte, error) {
	type Data struct {
		GUID string `json:"guid,omitempty"`
	}

	type OrgData struct {
		Data Data `json:"data,omitempty"`
	}

	type OrgRelationship struct {
		Org OrgData `json:"organization,omitempty"`
	}

	type ccDomain struct {
		GUID          string           `json:"guid,omitempty"`
		Name          string           `json:"name"`
		Internal      *bool            `json:"internal,omitempty"`
		Relationships *OrgRelationship `json:"relationships,omitempty"`
	}

	ccDom := ccDomain{
		Name: d.Name,
	}

	if d.Internal.IsSet {
		ccDom.Internal = &d.Internal.Value
	}

	if d.GUID != "" {
		ccDom.GUID = d.GUID
	}

	if d.OrganizationGUID != "" {
		ccDom.Relationships = &OrgRelationship{OrgData{Data{GUID: d.OrganizationGUID}}}
	}
	return json.Marshal(ccDom)
}

func (d *Domain) UnmarshalJSON(data []byte) error {
	var ccRouteStruct struct {
		GUID          string         `json:"guid,omitempty"`
		Name          string         `json:"name"`
		Internal      types.NullBool `json:"internal,omitempty"`
		Relationships struct {
			Organization struct {
				Data struct {
					GUID string `json:"guid,omitempty"`
				} `json:"data,omitempty"`
			} `json:"organization,omitempty"`
		} `json:"relationships,omitempty"`
		Metadata *Metadata
	}

	err := cloudcontroller.DecodeJSON(data, &ccRouteStruct)
	if err != nil {
		return err
	}

	d.GUID = ccRouteStruct.GUID
	d.Name = ccRouteStruct.Name
	d.Internal = ccRouteStruct.Internal
	d.OrganizationGUID = ccRouteStruct.Relationships.Organization.Data.GUID
	d.Metadata = ccRouteStruct.Metadata
	return nil
}

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

	_, warnings, err := client.makeRequest(requestParams{
		RequestName:  internal.GetDomainRouteReservationsRequest,
		URIParams:    internal.Params{"domain_guid": domainGUID},
		Query:        query,
		ResponseBody: &responseBody,
	})

	return responseBody.MatchingRoute, warnings, err
}

func (client Client) CreateDomain(domain Domain) (Domain, Warnings, error) {
	var responseBody Domain

	_, warnings, err := client.makeRequest(requestParams{
		RequestName:  internal.PostDomainRequest,
		RequestBody:  domain,
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

func (client Client) DeleteDomain(domainGUID string) (JobURL, Warnings, error) {
	jobURL, warnings, err := client.makeRequest(requestParams{
		RequestName: internal.DeleteDomainRequest,
		URIParams:   internal.Params{"domain_guid": domainGUID},
	})

	return jobURL, warnings, err
}

// GetDomain returns a domain with the given GUID.
func (client *Client) GetDomain(domainGUID string) (Domain, Warnings, error) {
	var responseBody Domain

	_, warnings, err := client.makeRequest(requestParams{
		RequestName:  internal.GetDomainRequest,
		URIParams:    internal.Params{"domain_guid": domainGUID},
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

func (client Client) GetDomains(query ...Query) ([]Domain, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetDomainsRequest,
		Query:       query,
	})
	if err != nil {
		return nil, nil, err
	}

	var fullDomainsList []Domain
	warnings, err := client.paginate(request, Domain{}, func(item interface{}) error {
		if domain, ok := item.(Domain); ok {
			fullDomainsList = append(fullDomainsList, domain)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   Domain{},
				Unexpected: item,
			}
		}
		return nil
	})

	return fullDomainsList, warnings, err
}

func (client Client) GetOrganizationDomains(orgGUID string, query ...Query) ([]Domain, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		URIParams:   internal.Params{"organization_guid": orgGUID},
		RequestName: internal.GetOrganizationDomainsRequest,
		Query:       query,
	})
	if err != nil {
		return nil, nil, err
	}

	var fullDomainsList []Domain
	warnings, err := client.paginate(request, Domain{}, func(item interface{}) error {
		if domain, ok := item.(Domain); ok {
			fullDomainsList = append(fullDomainsList, domain)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   Domain{},
				Unexpected: item,
			}
		}
		return nil
	})

	return fullDomainsList, warnings, err
}

func (client Client) SharePrivateDomainToOrgs(domainGuid string, sharedOrgs SharedOrgs) (Warnings, error) {
	_, warnings, err := client.makeRequest(requestParams{
		RequestName: internal.SharePrivateDomainRequest,
		URIParams:   internal.Params{"domain_guid": domainGuid},
		RequestBody: sharedOrgs,
	})

	return warnings, err
}

func (client Client) UnsharePrivateDomainFromOrg(domainGuid string, orgGUID string) (Warnings, error) {
	_, warnings, err := client.makeRequest(requestParams{
		RequestName: internal.DeleteSharedOrgFromDomainRequest,
		URIParams:   internal.Params{"domain_guid": domainGuid, "org_guid": orgGUID},
	})

	return warnings, err
}
