package ccv2

import (
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/internal"
)

// Organization represents a Cloud Controller Organization.
type Organization struct {
	GUID string
}

// UnmarshalJSON helps unmarshal a Cloud Controller Organization response.
func (org *Organization) UnmarshalJSON(data []byte) error {
	var ccOrg struct {
		Metadata internal.Metadata `json:"metadata"`
	}
	if err := json.Unmarshal(data, &ccOrg); err != nil {
		return err
	}

	org.GUID = ccOrg.Metadata.GUID
	return nil
}

// GetOrganizations returns back a list of Organizations based off of the
// provided queries.
func (client *Client) GetOrganizations(queries []Query) ([]Organization, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.OrganizationsRequest,
		Query:       FormatQueryParameters(queries),
	})
	if err != nil {
		return nil, nil, err
	}

	var fullOrgsList []Organization
	warnings, err := client.paginate(request, Organization{}, func(item interface{}) error {
		if app, ok := item.(Organization); ok {
			fullOrgsList = append(fullOrgsList, app)
		} else {
			return cloudcontroller.UnknownObjectInListError{
				Expected:   Organization{},
				Unexpected: item,
			}
		}
		return nil
	})

	return fullOrgsList, warnings, err
}
