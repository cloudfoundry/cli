package ccv2

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/internal"
)

// SpaceQuota represents the Cloud Controller configured quota assigned to the
// space.
type SpaceQuota struct {

	// GUID is the unique space quota identifier.
	GUID string

	// Name is the name given to the space quota.
	Name string
}

// UnmarshalJSON helps unmarshal a Cloud Controller Space Quota response.
func (spaceQuota *SpaceQuota) UnmarshalJSON(data []byte) error {
	var ccSpaceQuota struct {
		Metadata internal.Metadata `json:"metadata"`
		Entity   struct {
			Name string `json:"name"`
		} `json:"entity"`
	}
	err := cloudcontroller.DecodeJSON(data, &ccSpaceQuota)
	if err != nil {
		return err
	}

	spaceQuota.GUID = ccSpaceQuota.Metadata.GUID
	spaceQuota.Name = ccSpaceQuota.Entity.Name
	return nil
}

// GetSpaceQuotaDefinition returns a Space Quota.
func (client *Client) GetSpaceQuotaDefinition(guid string) (SpaceQuota, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetSpaceQuotaDefinitionRequest,
		URIParams:   Params{"space_quota_guid": guid},
	})
	if err != nil {
		return SpaceQuota{}, nil, err
	}

	var spaceQuota SpaceQuota
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &spaceQuota,
	}

	err = client.connection.Make(request, &response)
	return spaceQuota, response.Warnings, err
}

// GetSpaceQuotas returns all the space quotas for the org
func (client *Client) GetSpaceQuotas(orgGUID string) ([]SpaceQuota, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetOrganizationSpaceQuotasRequest,
		URIParams:   Params{"organization_guid": orgGUID},
	})

	if err != nil {
		return nil, nil, err
	}

	var spaceQuotas []SpaceQuota
	warnings, err := client.paginate(request, SpaceQuota{}, func(item interface{}) error {
		if spaceQuota, ok := item.(SpaceQuota); ok {
			spaceQuotas = append(spaceQuotas, spaceQuota)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   SpaceQuota{},
				Unexpected: item,
			}
		}
		return nil
	})

	return spaceQuotas, warnings, err
}

// SetSpaceQuota should set the quota for the space and returns the warnings
func (client *Client) SetSpaceQuota(spaceGUID string, quotaGUID string) (Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PutSpaceQuotaRequest,
		URIParams:   Params{"space_quota_guid": quotaGUID, "space_guid": spaceGUID},
	})

	if err != nil {
		return nil, err
	}

	response := cloudcontroller.Response{}
	err = client.connection.Make(request, &response)

	return response.Warnings, err
}
