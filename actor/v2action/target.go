package v2action

import "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"

type TargetSettings ccv2.TargetSettings

// SetTarget targets the Cloud Controller using the client and sets target
// information in the actor based on the response.
func (actor Actor) SetTarget(config Config, settings TargetSettings) (Warnings, error) {
	if config.Target() == settings.URL && config.SkipSSLValidation() == settings.SkipSSLValidation {
		return nil, nil
	}

	warnings, err := actor.CloudControllerClient.TargetCF(ccv2.TargetSettings(settings))
	if err != nil {
		return Warnings(warnings), err
	}

	config.SetTargetInformation(
		actor.CloudControllerClient.API(),
		actor.CloudControllerClient.APIVersion(),
		actor.CloudControllerClient.AuthorizationEndpoint(),
		actor.CloudControllerClient.MinCLIVersion(),
		actor.CloudControllerClient.DopplerEndpoint(),
		actor.CloudControllerClient.RoutingEndpoint(),
		settings.SkipSSLValidation,
	)
	config.SetTokenInformation("", "", "")

	return Warnings(warnings), nil
}

// ClearTarget clears target information from the actor.
func (Actor) ClearTarget(config Config) {
	config.SetTargetInformation("", "", "", "", "", "", false)
	config.SetTokenInformation("", "", "")
}

// ClearTarget clears the targeted org and space in the config.
func (Actor) ClearOrganizationAndSpace(config Config) {
	config.UnsetOrganizationInformation()
	config.UnsetSpaceInformation()
}
