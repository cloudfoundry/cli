package ccv2

import (
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/internal"
)

// Organization represents a Cloud Controller Organization.
type Organization struct {
	GUID                        string
	Name                        string
	QuotaDefinitionGUID         string
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
	if err := json.Unmarshal(data, &ccOrg); err != nil {
		return err
	}

	org.GUID = ccOrg.Metadata.GUID
	org.Name = ccOrg.Entity.Name
	org.QuotaDefinitionGUID = ccOrg.Entity.QuotaDefinitionGUID
	org.DefaultIsolationSegmentGUID = ccOrg.Entity.DefaultIsolationSegmentGUID
	return nil
}

//go:generate go run $GOPATH/src/code.cloudfoundry.org/cli/util/codegen/generate.go Organization codetemplates/delete_async_by_guid.go.template delete_organization.go
//go:generate go run $GOPATH/src/code.cloudfoundry.org/cli/util/codegen/generate.go Organization codetemplates/delete_async_by_guid_test.go.template delete_organization_test.go

// GetOrganization returns an Organization associated with the provided guid.
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
// provided queries.
func (client *Client) GetOrganizations(queries ...QQuery) ([]Organization, Warnings, error) {
	allQueries := FormatQueryParameters(queries)
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
