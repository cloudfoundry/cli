package ccv3

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
)

// Represents a Cloud Controller V3 Feature Flag.
type FeatureFlag struct {
	Name string
}

// Lists feature flags.
func (client *Client) GetFeatureFlags() ([]FeatureFlag, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetFeatureFlagsRequest,
	})

	if err != nil {
		return nil, nil, err
	}

	var fullFeatureFlagList []FeatureFlag
	warnings, err := client.paginate(request, FeatureFlag{}, func(item interface{}) error {
		if featureFlag, ok := item.(FeatureFlag); ok {
			fullFeatureFlagList = append(fullFeatureFlagList, featureFlag)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   FeatureFlag{},
				Unexpected: item,
			}
		}
		return nil
	})

	return fullFeatureFlagList, warnings, err

}
