package ccv3

import (
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
)

// Organization represents a Cloud Controller V3 Organization.
type Organization struct {
	// GUID is the unique organization identifier.
	GUID string `json:"guid,omitempty"`
	// Name is the name of the organization.
	Name string `json:"name"`
	// QuotaGUID is the GUID of the organization Quota applied to this Organization
	QuotaGUID string `json:"-"`

	// Metadata is used for custom tagging of API resources
	Metadata *Metadata `json:"metadata,omitempty"`
}

func (org *Organization) UnmarshalJSON(data []byte) error {
	type alias Organization
	var aliasOrg alias
	err := json.Unmarshal(data, &aliasOrg)
	if err != nil {
		return err
	}

	*org = Organization(aliasOrg)

	remainingFields := new(struct {
		Relationships struct {
			Quota struct {
				Data struct {
					GUID string
				}
			}
		}
	})

	err = json.Unmarshal(data, &remainingFields)
	if err != nil {
		return err
	}

	org.QuotaGUID = remainingFields.Relationships.Quota.Data.GUID

	return nil
}

// CreateOrganization creates an organization with the given name.
func (client *Client) CreateOrganization(orgName string) (Organization, Warnings, error) {
	org := Organization{Name: orgName}
	var responseBody Organization

	_, warnings, err := client.makeRequest(requestParams{
		RequestName:  internal.PostOrganizationRequest,
		RequestBody:  org,
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

// DeleteOrganization deletes the organization with the given GUID.
func (client *Client) DeleteOrganization(orgGUID string) (JobURL, Warnings, error) {
	jobURL, warnings, err := client.makeRequest(requestParams{
		RequestName: internal.DeleteOrganizationRequest,
		URIParams:   internal.Params{"organization_guid": orgGUID},
	})

	return jobURL, warnings, err
}

// GetDefaultDomain gets the default domain for the organization with the given GUID.
func (client *Client) GetDefaultDomain(orgGUID string) (Domain, Warnings, error) {
	var responseBody Domain

	_, warnings, err := client.makeRequest(requestParams{
		RequestName:  internal.GetDefaultDomainRequest,
		URIParams:    internal.Params{"organization_guid": orgGUID},
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

// GetIsolationSegmentOrganizations lists organizations
// entitled to an isolation segment.
func (client *Client) GetIsolationSegmentOrganizations(isolationSegmentGUID string) ([]Organization, Warnings, error) {
	var resources []Organization

	_, warnings, err := client.makeListRequest(requestParams{
		RequestName:  internal.GetIsolationSegmentOrganizationsRequest,
		URIParams:    internal.Params{"isolation_segment_guid": isolationSegmentGUID},
		ResponseBody: Organization{},
		AppendToList: func(item interface{}) error {
			resources = append(resources, item.(Organization))
			return nil
		},
	})

	return resources, warnings, err
}

// GetOrganization gets an organization by the given guid.
func (client *Client) GetOrganization(orgGUID string) (Organization, Warnings, error) {
	var responseBody Organization

	_, warnings, err := client.makeRequest(requestParams{
		RequestName:  internal.GetOrganizationRequest,
		URIParams:    internal.Params{"organization_guid": orgGUID},
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

// GetOrganizations lists organizations with optional filters.
func (client *Client) GetOrganizations(query ...Query) ([]Organization, Warnings, error) {
	var resources []Organization

	_, warnings, err := client.makeListRequest(requestParams{
		RequestName:  internal.GetOrganizationsRequest,
		Query:        query,
		ResponseBody: Organization{},
		AppendToList: func(item interface{}) error {
			resources = append(resources, item.(Organization))
			return nil
		},
	})

	return resources, warnings, err
}

// UpdateOrganization updates an organization with the given properties.
func (client *Client) UpdateOrganization(org Organization) (Organization, Warnings, error) {
	orgGUID := org.GUID
	org.GUID = ""

	var responseBody Organization

	_, warnings, err := client.makeRequest(requestParams{
		RequestName:  internal.PatchOrganizationRequest,
		URIParams:    internal.Params{"organization_guid": orgGUID},
		RequestBody:  org,
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}
