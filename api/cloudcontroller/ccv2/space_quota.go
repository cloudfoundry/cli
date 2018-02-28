package ccv2

import (
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
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
	if err := json.Unmarshal(data, &ccSpaceQuota); err != nil {
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
		Result: &spaceQuota,
	}

	err = client.connection.Make(request, &response)
	return spaceQuota, response.Warnings, err
}
