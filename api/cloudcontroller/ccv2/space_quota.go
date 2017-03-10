package ccv2

import (
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/internal"
)

type SpaceQuota struct {
	GUID string
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

// GetSpaceQuota returns a Space Quota.
func (client *Client) GetSpaceQuota(guid string) (SpaceQuota, Warnings, error) {
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
