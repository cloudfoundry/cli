package ccv3

import (
	"bytes"
	"fmt"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
)

type SpaceFeature struct {
	Name    string
	Enabled bool
}

func (client *Client) GetSpaceFeature(spaceGUID string, featureName string) (bool, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetSpaceFeatureRequest,
		URIParams:   map[string]string{"space_guid": spaceGUID, "feature": featureName},
	})
	if err != nil {
		return false, nil, err
	}

	var feature SpaceFeature
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &feature,
	}
	err = client.connection.Make(request, &response)

	return feature.Enabled, response.Warnings, err
}

func (client *Client) UpdateSpaceFeature(spaceGUID string, enabled bool, featureName string) (Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PatchSpaceFeaturesRequest,
		Body:        bytes.NewReader([]byte(fmt.Sprintf(`{"enabled":%t}`, enabled))),
		URIParams:   map[string]string{"space_guid": spaceGUID, "feature": featureName},
	})
	if err != nil {
		return nil, err
	}

	response := cloudcontroller.Response{}
	err = client.connection.Make(request, &response)

	return response.Warnings, err
}
