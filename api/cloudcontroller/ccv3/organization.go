package ccv3

import (
	"bytes"
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
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
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetIsolationSegmentOrganizationsRequest,
		URIParams:   map[string]string{"isolation_segment_guid": isolationSegmentGUID},
	})
	if err != nil {
		return nil, nil, err
	}

	var fullOrgsList []Organization
	warnings, err := client.paginate(request, Organization{}, func(item interface{}) error {
		if app, ok := item.(Organization); ok {
			fullOrgsList = append(fullOrgsList, app)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   Organization{},
				Unexpected: item,
			}
		}
		return nil
	})

	return fullOrgsList, warnings, err
}

// GetOrganization gets an organization by the given guid.
func (client *Client) GetOrganization(orgGUID string) (Organization, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetOrganizationRequest,
		URIParams:   map[string]string{"organization_guid": orgGUID},
	})

	if err != nil {
		return Organization{}, nil, err
	}

	var responseOrg Organization
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &responseOrg,
	}
	err = client.connection.Make(request, &response)
	if err != nil {
		return Organization{}, response.Warnings, err
	}

	return responseOrg, response.Warnings, nil
}

// GetOrganizations lists organizations with optional filters.
func (client *Client) GetOrganizations(query ...Query) ([]Organization, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetOrganizationsRequest,
		Query:       query,
	})
	if err != nil {
		return nil, nil, err
	}

	var fullOrgsList []Organization
	warnings, err := client.paginate(request, Organization{}, func(item interface{}) error {
		if app, ok := item.(Organization); ok {
			fullOrgsList = append(fullOrgsList, app)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   Organization{},
				Unexpected: item,
			}
		}
		return nil
	})

	return fullOrgsList, warnings, err
}

// UpdateOrganization updates an organization with the given properties.
func (client *Client) UpdateOrganization(org Organization) (Organization, Warnings, error) {
	orgGUID := org.GUID
	org.GUID = ""
	orgBytes, err := json.Marshal(org)
	if err != nil {
		return Organization{}, nil, err
	}
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PatchOrganizationRequest,
		Body:        bytes.NewReader(orgBytes),
		URIParams:   map[string]string{"organization_guid": orgGUID},
	})

	if err != nil {
		return Organization{}, nil, err
	}

	var responseOrg Organization
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &responseOrg,
	}
	err = client.connection.Make(request, &response)

	if err != nil {
		return Organization{}, nil, err
	}
	return responseOrg, nil, err
}
