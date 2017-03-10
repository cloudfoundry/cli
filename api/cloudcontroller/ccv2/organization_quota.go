package ccv2

import (
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/internal"
)

// OrganizationQuota is the definition of a quota for an organization.
type OrganizationQuota struct {
	GUID string
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
	if err := json.Unmarshal(data, &ccOrgQuota); err != nil {
		return err
	}

	application.GUID = ccOrgQuota.Metadata.GUID
	application.Name = ccOrgQuota.Entity.Name

	return nil
}

// GetOrganizaitonQuota gets an organization quota (quota definition) from the API.
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
		Result: &orgQuota,
	}

	err = client.connection.Make(request, &response)
	return orgQuota, response.Warnings, err
}
