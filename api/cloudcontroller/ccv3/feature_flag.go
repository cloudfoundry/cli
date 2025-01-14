package ccv3

import (
	ccv3internal "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/api/internal"
	"code.cloudfoundry.org/cli/resources"
)

func (client *Client) GetFeatureFlag(flagName string) (resources.FeatureFlag, Warnings, error) {
	var responseBody resources.FeatureFlag

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  ccv3internal.GetFeatureFlagRequest,
		URIParams:    internal.Params{"name": flagName},
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

// GetFeatureFlags lists feature flags.
func (client *Client) GetFeatureFlags() ([]resources.FeatureFlag, Warnings, error) {
	var featureFlags []resources.FeatureFlag

	_, warnings, err := client.MakeListRequest(RequestParams{
		RequestName:  ccv3internal.GetFeatureFlagsRequest,
		ResponseBody: resources.FeatureFlag{},
		AppendToList: func(item interface{}) error {
			featureFlags = append(featureFlags, item.(resources.FeatureFlag))
			return nil
		},
	})

	return featureFlags, warnings, err
}

func (client *Client) UpdateFeatureFlag(flag resources.FeatureFlag) (resources.FeatureFlag, Warnings, error) {
	var responseBody resources.FeatureFlag

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  ccv3internal.PatchFeatureFlagRequest,
		URIParams:    internal.Params{"name": flag.Name},
		RequestBody:  flag,
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}
