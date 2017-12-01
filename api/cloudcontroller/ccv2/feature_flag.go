package ccv2

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/internal"
)

// FeatureFlag represents a Cloud Controller feature flag.
type FeatureFlag struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

func (client Client) GetConfigFeatureFlags() ([]FeatureFlag, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetConfigFeatureFlagsRequest,
	})
	if err != nil {
		return nil, nil, err
	}

	var featureFlags []FeatureFlag
	response := cloudcontroller.Response{
		Result: &featureFlags,
	}

	err = client.connection.Make(request, &response)
	return featureFlags, response.Warnings, err
}
