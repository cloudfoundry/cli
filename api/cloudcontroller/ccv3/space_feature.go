package ccv3

import (
	ccv3internal "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/api/internal"
	"code.cloudfoundry.org/cli/resources"
)

func (client *Client) GetSpaceFeature(spaceGUID string, featureName string) (bool, Warnings, error) {
	var responseBody resources.SpaceFeature

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  ccv3internal.GetSpaceFeatureRequest,
		URIParams:    internal.Params{"space_guid": spaceGUID, "feature": featureName},
		ResponseBody: &responseBody,
	})

	return responseBody.Enabled, warnings, err
}

func (client *Client) UpdateSpaceFeature(spaceGUID string, enabled bool, featureName string) (Warnings, error) {
	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName: ccv3internal.PatchSpaceFeaturesRequest,
		URIParams:   internal.Params{"space_guid": spaceGUID, "feature": featureName},
		RequestBody: struct {
			Enabled bool `json:"enabled"`
		}{Enabled: enabled},
	})

	return warnings, err
}
