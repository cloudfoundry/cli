package ccv3

import (
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/api/cloudcontroller/jsonry"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/types"
)

type Domain struct {
	GUID             string         `json:"guid,omitempty"`
	Name             string         `json:"name"`
	Internal         types.NullBool `json:"internal,omitempty"`
	OrganizationGUID string         `jsonry:"relationships.organization.data.guid,omitempty"`
	RouterGroup      string         `jsonry:"router_group.guid,omitempty"`

	// Metadata is used for custom tagging of API resources
	Metadata *resources.Metadata `json:"metadata,omitempty"`
}

func (d Domain) MarshalJSON() ([]byte, error) {
	type domainWithBoolPointer struct {
		GUID             string `jsonry:"guid,omitempty"`
		Name             string `jsonry:"name"`
		Internal         *bool  `jsonry:"internal,omitempty"`
		OrganizationGUID string `jsonry:"relationships.organization.data.guid,omitempty"`
		RouterGroup      string `jsonry:"router_group.guid,omitempty"`
	}

	clone := domainWithBoolPointer{
		GUID:             d.GUID,
		Name:             d.Name,
		OrganizationGUID: d.OrganizationGUID,
		RouterGroup:      d.RouterGroup,
	}

	if d.Internal.IsSet {
		clone.Internal = &d.Internal.Value
	}
	return jsonry.Marshal(clone)
}

func (d *Domain) UnmarshalJSON(data []byte) error {
	type alias Domain
	var defaultUnmarshalledDomain alias
	err := jsonry.Unmarshal(data, &defaultUnmarshalledDomain)
	if err != nil {
		return err
	}

	*d = Domain(defaultUnmarshalledDomain)

	type RemainingFieldsStruct struct {
		Internal types.NullBool `json:"internal,omitempty"`
	}

	var remainingFields RemainingFieldsStruct
	err = json.Unmarshal(data, &remainingFields)
	if err != nil {
		return err
	}

	d.Internal = remainingFields.Internal

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

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.GetDomainRouteReservationsRequest,
		URIParams:    internal.Params{"domain_guid": domainGUID},
		Query:        query,
		ResponseBody: &responseBody,
	})

	return responseBody.MatchingRoute, warnings, err
}

func (client Client) CreateDomain(domain Domain) (Domain, Warnings, error) {
	var responseBody Domain

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
func (client *Client) GetDomain(domainGUID string) (Domain, Warnings, error) {
	var responseBody Domain

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.GetDomainRequest,
		URIParams:    internal.Params{"domain_guid": domainGUID},
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

func (client Client) GetDomains(query ...Query) ([]Domain, Warnings, error) {
	var resources []Domain

	_, warnings, err := client.MakeListRequest(RequestParams{
		RequestName:  internal.GetDomainsRequest,
		Query:        query,
		ResponseBody: Domain{},
		AppendToList: func(item interface{}) error {
			resources = append(resources, item.(Domain))
			return nil
		},
	})

	return resources, warnings, err
}

func (client Client) GetOrganizationDomains(orgGUID string, query ...Query) ([]Domain, Warnings, error) {
	var resources []Domain

	_, warnings, err := client.MakeListRequest(RequestParams{
		URIParams:    internal.Params{"organization_guid": orgGUID},
		RequestName:  internal.GetOrganizationDomainsRequest,
		Query:        query,
		ResponseBody: Domain{},
		AppendToList: func(item interface{}) error {
			resources = append(resources, item.(Domain))
			return nil
		},
	})

	return resources, warnings, err
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
