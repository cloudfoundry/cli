package v2action

import "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"

type TargetSettings ccv2.TargetSettings

// ClearTarget clears target information from the actor.
func (actor Actor) ClearTarget() {
	actor.Config.SetTargetInformation("", "", "", "", "", "", false)
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

	actor.Config.SetTargetInformation(
		actor.CloudControllerClient.API(),
		actor.CloudControllerClient.APIVersion(),
		actor.CloudControllerClient.AuthorizationEndpoint(),
		actor.CloudControllerClient.MinCLIVersion(),
		actor.CloudControllerClient.DopplerEndpoint(),
		actor.CloudControllerClient.RoutingEndpoint(),
		settings.SkipSSLValidation,
	)
	actor.Config.SetTokenInformation("", "", "")

	return Warnings(warnings), nil
}
