package v2action

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"

	"code.cloudfoundry.org/cli/util/configv3"
)

type TargetSettings ccv2.TargetSettings

// ClearTarget clears target information from the actor.
func (actor Actor) ClearTarget() {
	actor.Config.SetTargetInformation(configv3.TargetInformationArgs{})
	actor.Config.SetTokenInformation("", "", "")
}

// MinCLIVersion returns the minimum CLI version that the Cloud Controller
// requires.
func (actor Actor) MinCLIVersion() string {
	return actor.CloudControllerClient.MinCLIVersion()
}

// SetTarget targets the Cloud Controller using the client and sets target
// information in the actor based on the response.
func (actor Actor) SetTarget(settings TargetSettings) (Warnings, error) {
	if actor.Config.Target() == settings.URL && actor.Config.SkipSSLValidation() == settings.SkipSSLValidation {
		return nil, nil
	}

	warnings, err := actor.CloudControllerClient.TargetCF(ccv2.TargetSettings(settings))
	if err != nil {
		return Warnings(warnings), err
	}

	actor.Config.SetTargetInformation(configv3.TargetInformationArgs{
		Api:               actor.CloudControllerClient.API(),
		ApiVersion:        actor.CloudControllerClient.APIVersion(),
		Auth:              actor.CloudControllerClient.AuthorizationEndpoint(),
		MinCLIVersion:     actor.CloudControllerClient.MinCLIVersion(),
		Doppler:           actor.CloudControllerClient.DopplerEndpoint(),
		Routing:           actor.CloudControllerClient.RoutingEndpoint(),
		SkipSSLValidation: settings.SkipSSLValidation,
	})
	actor.Config.SetTokenInformation("", "", "")

	return Warnings(warnings), nil
}

func (actor Actor) AuthorizationEndpoint() string {
	return actor.CloudControllerClient.AuthorizationEndpoint()
}
