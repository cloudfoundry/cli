package ccv3

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/resources"
)

func (client *Client) CreateOrganization(orgName string) (resources.Organization, Warnings, error) {
	org := resources.Organization{Name: orgName}
	var responseBody resources.Organization

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.PostOrganizationRequest,
		RequestBody:  org,
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

// DeleteOrganization deletes the organization with the given GUID.
func (client *Client) DeleteOrganization(orgGUID string) (JobURL, Warnings, error) {
	jobURL, warnings, err := client.MakeRequest(RequestParams{
		RequestName: internal.DeleteOrganizationRequest,
		URIParams:   internal.Params{"organization_guid": orgGUID},
	})

	return jobURL, warnings, err
}

// GetDefaultDomain gets the default domain for the organization with the given GUID.
func (client *Client) GetDefaultDomain(orgGUID string) (resources.Domain, Warnings, error) {
	var responseBody resources.Domain

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.GetDefaultDomainRequest,
		URIParams:    internal.Params{"organization_guid": orgGUID},
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

// GetIsolationSegmentOrganizations lists organizations
// entitled to an isolation segment.
func (client *Client) GetIsolationSegmentOrganizations(isolationSegmentGUID string) ([]resources.Organization, Warnings, error) {
	var organizations []resources.Organization

	_, warnings, err := client.MakeListRequest(RequestParams{
		RequestName:  internal.GetIsolationSegmentOrganizationsRequest,
		URIParams:    internal.Params{"isolation_segment_guid": isolationSegmentGUID},
		ResponseBody: resources.Organization{},
		AppendToList: func(item interface{}) error {
			organizations = append(organizations, item.(resources.Organization))
			return nil
		},
	})

	return organizations, warnings, err
}

// GetOrganization gets an organization by the given guid.
func (client *Client) GetOrganization(orgGUID string) (resources.Organization, Warnings, error) {
	var responseBody resources.Organization

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.GetOrganizationRequest,
		URIParams:    internal.Params{"organization_guid": orgGUID},
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

// GetOrganizations lists organizations with optional filters.
func (client *Client) GetOrganizations(query ...Query) ([]resources.Organization, Warnings, error) {
	var organizations []resources.Organization

	_, warnings, err := client.MakeListRequest(RequestParams{
		RequestName:  internal.GetOrganizationsRequest,
		Query:        query,
		ResponseBody: resources.Organization{},
		AppendToList: func(item interface{}) error {
			organizations = append(organizations, item.(resources.Organization))
			return nil
		},
	})

	return organizations, warnings, err
}

// UpdateOrganization updates an organization with the given properties.
func (client *Client) UpdateOrganization(org resources.Organization) (resources.Organization, Warnings, error) {
	orgGUID := org.GUID
	org.GUID = ""

	var responseBody resources.Organization

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.PatchOrganizationRequest,
		URIParams:    internal.Params{"organization_guid": orgGUID},
		RequestBody:  org,
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}
