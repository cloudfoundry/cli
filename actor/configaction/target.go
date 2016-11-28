package configaction

import "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"

type TargetSettings ccv2.TargetSettings

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
		actor.CloudControllerClient.LoggregatorEndpoint(),
		actor.CloudControllerClient.DopplerEndpoint(),
		actor.CloudControllerClient.TokenEndpoint(),
		actor.CloudControllerClient.RoutingEndpoint(),
		settings.SkipSSLValidation,
	)
	actor.Config.SetTokenInformation("", "", "")

	return Warnings(warnings), nil
}

func (actor Actor) ClearTarget() {
	actor.Config.SetTargetInformation("", "", "", "", "", "", "", false)
	actor.Config.SetTokenInformation("", "", "")
}
