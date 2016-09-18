package configactions

func (actor Actor) SetTarget(CCAPI string, skipSSLValidation bool) (Warnings, error) {
	warnings, err := actor.CloudControllerClient.TargetCF(CCAPI, skipSSLValidation)
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
		skipSSLValidation,
	)

	return Warnings(warnings), nil
}

func (actor Actor) ClearTarget() {
	actor.Config.SetTargetInformation("", "", "", "", "", "", "", false)
	actor.Config.SetTokenInformation("", "", "")
}
