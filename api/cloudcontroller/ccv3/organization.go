package ccv3

import (
	"bytes"
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/types"
)

// Organization represents a Cloud Controller V3 Organization.
type Organization struct {
	// GUID is the unique organization identifier.
	GUID string `json:"guid,omitempty"`
	// Name is the name of the organization.
	Name string `json:"name"`

	// Metadata is used for custom tagging of API resources
	Metadata struct {
		Labels map[string]types.NullString `json:"labels,omitempty"`
	} `json:"metadata,omitempty"`
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
