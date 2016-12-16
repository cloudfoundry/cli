package ccv2

import (
	"encoding/json"
	"net/http"
	"net/url"

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

	fullOrgsList := []Organization{}
	fullWarningsList := Warnings{}

	for {
		var orgs []Organization
		wrapper := PaginatedWrapper{
			Resources: &orgs,
		}
		response := cloudcontroller.Response{
			Result: &wrapper,
		}

		err = client.connection.Make(request, &response)
		fullWarningsList = append(fullWarningsList, response.Warnings...)
		if err != nil {
			return nil, fullWarningsList, err
		}

		fullOrgsList = append(fullOrgsList, orgs...)

		if wrapper.NextURL == "" {
			break
		}

		request, err = client.newHTTPRequest(requestOptions{
			URI:    wrapper.NextURL,
			Method: http.MethodGet,
		})
		if err != nil {
			return nil, fullWarningsList, err
		}
	}

	return fullOrgsList, fullWarningsList, nil
}

// DeleteOrganization deletes the Organization associated with the provided
// GUID.
func (client *Client) DeleteOrganization(orgGUID string) (Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.DeleteOrganizationRequest,
		URIParams:   map[string]string{"organization_guid": orgGUID},
		Query: url.Values{
			"recursive": {"true"},
			"async":     {"true"},
		},
	})
	if err != nil {
		return nil, err
	}

	var response cloudcontroller.Response
	err = client.connection.Make(request, &response)
	return response.Warnings, err
}
