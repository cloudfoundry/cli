package ccv3

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
)

// Represents a Cloud Controller V3 Feature Flag.
type FeatureFlag struct {
	Name               string `json:"name"`
	Enabled            bool   `json:"enabled"`
	CustomErrorMessage string `json:"custom_error_message"`
}

func (client *Client) GetFeatureFlag(flagName string) (FeatureFlag, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetFeatureFlagRequest,
		URIParams:   map[string]string{"name": flagName},
	})
	if err != nil {
		return FeatureFlag{}, nil, err
	}
	var ccFlag FeatureFlag
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &ccFlag,
	}

	err = client.connection.Make(request, &response)

	if err != nil {
		switch response.HTTPResponse.StatusCode {
		case 404:
			return FeatureFlag{}, response.Warnings, ccerror.FeatureFlagNotFoundError{}
		default:
			return FeatureFlag{}, response.Warnings, err
		}
	}
	return ccFlag, response.Warnings, nil
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
