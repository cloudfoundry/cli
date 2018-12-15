package ccv2

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/internal"
)

// OrganizationQuota is the definition of a quota for an organization.
type OrganizationQuota struct {

	// GUID is the unique OrganizationQuota identifier.
	GUID string

	// Name is the name of the OrganizationQuota.
	Name string
}

// UnmarshalJSON helps unmarshal a Cloud Controller organization quota response.
func (application *OrganizationQuota) UnmarshalJSON(data []byte) error {
	var ccOrgQuota struct {
		Metadata internal.Metadata `json:"metadata"`
		Entity   struct {
			Name string `json:"name"`
		} `json:"entity"`
	}
	err := cloudcontroller.DecodeJSON(data, &ccOrgQuota)
	if err != nil {
		return err
	}

	application.GUID = ccOrgQuota.Metadata.GUID
	application.Name = ccOrgQuota.Entity.Name

	return nil
}

// GetOrganizationQuota returns an Organization Quota associated with the
// provided GUID.
func (client *Client) GetOrganizationQuota(guid string) (OrganizationQuota, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetOrganizationQuotaDefinitionRequest,
		URIParams:   Params{"organization_quota_guid": guid},
	})
	if err != nil {
		return OrganizationQuota{}, nil, err
	}

	var orgQuota OrganizationQuota
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &orgQuota,
	}

	err = client.connection.Make(request, &response)
	return orgQuota, response.Warnings, err
}

// GetOrganizationQuotas returns an Organization Quota list associated with the
// provided filters.
func (client *Client) GetOrganizationQuotas(filters ...Filter) ([]OrganizationQuota, Warnings, error) {
	allQueries := ConvertFilterParameters(filters)
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetOrganizationQuotaDefinitionsRequest,
		Query:       allQueries,
	})

	if err != nil {
		return []OrganizationQuota{}, nil, err
	}

	var fullOrgQuotasList []OrganizationQuota

	warnings, err := client.paginate(request, OrganizationQuota{}, func(item interface{}) error {
		if org, ok := item.(OrganizationQuota); ok {
			fullOrgQuotasList = append(fullOrgQuotasList, org)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   OrganizationQuota{},
				Unexpected: item,
			}
		}
		return nil
	})

	return fullOrgQuotasList, warnings, err
}
