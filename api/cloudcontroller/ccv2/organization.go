package ccv2

import (
	"bytes"
	"encoding/json"
	"net/url"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/internal"
)

// Organization represents a Cloud Controller Organization.
type Organization struct {

	// GUID is the unique Organization identifier.
	GUID string

	// Name is the organization's name.
	Name string

	// QuotaDefinitionGUID is unique identifier of the quota assigned to this
	// organization.
	QuotaDefinitionGUID string

	// DefaultIsolationSegmentGUID is the unique identifier of the isolation
	// segment this organization is tagged with.
	DefaultIsolationSegmentGUID string
}

// UnmarshalJSON helps unmarshal a Cloud Controller Organization response.
func (org *Organization) UnmarshalJSON(data []byte) error {
	var ccOrg struct {
		Metadata internal.Metadata `json:"metadata"`
		Entity   struct {
			Name                        string `json:"name"`
			QuotaDefinitionGUID         string `json:"quota_definition_guid"`
			DefaultIsolationSegmentGUID string `json:"default_isolation_segment_guid"`
		} `json:"entity"`
	}
	err := cloudcontroller.DecodeJSON(data, &ccOrg)
	if err != nil {
		return err
	}

	org.GUID = ccOrg.Metadata.GUID
	org.Name = ccOrg.Entity.Name
	org.QuotaDefinitionGUID = ccOrg.Entity.QuotaDefinitionGUID
	org.DefaultIsolationSegmentGUID = ccOrg.Entity.DefaultIsolationSegmentGUID
	return nil
}

type createOrganizationRequestBody struct {
	Name                string `json:"name"`
	QuotaDefinitionGUID string `json:"quota_definition_guid,omitempty"`
}

func (client *Client) CreateOrganization(orgName string, quotaGUID string) (Organization, Warnings, error) {
	requestBody := createOrganizationRequestBody{
		Name:                orgName,
		QuotaDefinitionGUID: quotaGUID,
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return Organization{}, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PostOrganizationRequest,
		Body:        bytes.NewReader(bodyBytes),
	})

	if err != nil {
		return Organization{}, nil, err
	}

	var org Organization
	response := cloudcontroller.Response{
		Result: &org,
	}

	err = client.connection.Make(request, &response)
	return org, response.Warnings, err
}

// DeleteOrganization deletes the Organization associated with the provided
// GUID. It will return the Cloud Controller job that is assigned to the
// Organization deletion.
func (client *Client) DeleteOrganization(guid string) (Job, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.DeleteOrganizationRequest,
		URIParams:   Params{"organization_guid": guid},
		Query: url.Values{
			"recursive": {"true"},
			"async":     {"true"},
		},
	})
	if err != nil {
		return Job{}, nil, err
	}

	var job Job
	response := cloudcontroller.Response{
		Result: &job,
	}

	err = client.connection.Make(request, &response)
	return job, response.Warnings, err
}

// GetOrganization returns an Organization associated with the provided GUID.
func (client *Client) GetOrganization(guid string) (Organization, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetOrganizationRequest,
		URIParams:   Params{"organization_guid": guid},
	})
	if err != nil {
		return Organization{}, nil, err
	}

	var org Organization
	response := cloudcontroller.Response{
		Result: &org,
	}

	err = client.connection.Make(request, &response)
	return org, response.Warnings, err
}

// GetOrganizations returns back a list of Organizations based off of the
// provided filters.
func (client *Client) GetOrganizations(filters ...Filter) ([]Organization, Warnings, error) {
	allQueries := ConvertFilterParameters(filters)
	allQueries.Add("order-by", "name")
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetOrganizationsRequest,
		Query:       allQueries,
	})

	if err != nil {
		return nil, nil, err
	}

	var fullOrgsList []Organization
	warnings, err := client.paginate(request, Organization{}, func(item interface{}) error {
		if org, ok := item.(Organization); ok {
			fullOrgsList = append(fullOrgsList, org)
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

type updateOrgManagerByUsernameRequestBody struct {
	Username string `json:"username"`
}

// UpdateOrganizationManager assigns the org manager role to the UAA user or client with the provided ID.
func (client *Client) UpdateOrganizationManager(guid string, uaaID string) (Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PutOrganizationManagerRequest,
		URIParams:   Params{"organization_guid": guid, "manager_guid": uaaID},
	})
	if err != nil {
		return nil, err
	}

	response := cloudcontroller.Response{}
	err = client.connection.Make(request, &response)

	return response.Warnings, err
}

// UpdateOrganizationManagerByUsername assigns the org manager role to the user with the provided name.
func (client *Client) UpdateOrganizationManagerByUsername(guid string, username string) (Warnings, error) {
	requestBody := updateOrgManagerByUsernameRequestBody{
		Username: username,
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PutOrganizationManagerByUsernameRequest,
		Body:        bytes.NewReader(body),
		URIParams:   Params{"organization_guid": guid},
	})
	if err != nil {
		return nil, err
	}

	response := cloudcontroller.Response{}
	err = client.connection.Make(request, &response)

	return response.Warnings, err
}

type updateOrgUserByUsernameRequestBody struct {
	Username string `json:"username"`
}

func (client Client) UpdateOrganizationUserByUsername(orgGUID string, username string) (Warnings, error) {
	requestBody := updateOrgUserByUsernameRequestBody{
		Username: username,
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return Warnings{}, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PutOrganizationUserByUsernameRequest,
		Body:        bytes.NewReader(body),
		URIParams:   Params{"organization_guid": orgGUID},
	})
	if err != nil {
		return Warnings{}, err
	}

	response := cloudcontroller.Response{}
	err = client.connection.Make(request, &response)

	return response.Warnings, err
}
