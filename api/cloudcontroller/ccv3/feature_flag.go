package ccv3

import (
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
)

// FeatureFlag represents a Cloud Controller V3 Feature Flag.
type FeatureFlag struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

func (f FeatureFlag) MarshalJSON() ([]byte, error) {
	var ccBodyFlag struct {
		Enabled bool `json:"enabled"`
	}

	ccBodyFlag.Enabled = f.Enabled

	return json.Marshal(ccBodyFlag)
}

func (client *Client) GetFeatureFlag(flagName string) (FeatureFlag, Warnings, error) {
	var responseBody FeatureFlag

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.GetFeatureFlagRequest,
		URIParams:    internal.Params{"name": flagName},
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

// GetFeatureFlags lists feature flags.
func (client *Client) GetFeatureFlags() ([]FeatureFlag, Warnings, error) {
	var resources []FeatureFlag

	_, warnings, err := client.MakeListRequest(RequestParams{
		RequestName:  internal.GetFeatureFlagsRequest,
		ResponseBody: FeatureFlag{},
		AppendToList: func(item interface{}) error {
			resources = append(resources, item.(FeatureFlag))
			return nil
		},
	})

	return resources, warnings, err
}

func (client *Client) UpdateFeatureFlag(flag FeatureFlag) (FeatureFlag, Warnings, error) {
	var responseBody FeatureFlag

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.PatchFeatureFlagRequest,
		URIParams:    internal.Params{"name": flag.Name},
		RequestBody:  flag,
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}
