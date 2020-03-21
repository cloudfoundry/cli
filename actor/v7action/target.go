package v7action

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
)

type TargetSettings ccv3.TargetSettings

// SetTarget targets the Cloud Controller using the client and sets target
// information in the actor based on the response.
func (actor Actor) SetTarget(settings TargetSettings) (Warnings, error) {
	if actor.Config.Target() == settings.URL && actor.Config.SkipSSLValidation() == settings.SkipSSLValidation {
		return nil, nil
	}

	rootInfo, warnings, err := actor.CloudControllerClient.TargetCF(ccv3.TargetSettings(settings))
	if err != nil {
		return Warnings(warnings), err
	}

	actor.Config.SetTargetInformation(settings.URL,
		rootInfo.CloudControllerAPIVersion(),
		rootInfo.UAA(),
		"", // Oldest supported V3 version should be OK
		rootInfo.Logging(),
		rootInfo.Routing(),
		settings.SkipSSLValidation,
	)
	actor.Config.SetTokenInformation("", "", "")
	return Warnings(warnings), nil
}

// ClearTarget clears target information from the actor.
func (actor Actor) ClearTarget() {
	actor.Config.SetTargetInformation("", "", "", "", "", "", false)
	actor.Config.SetTokenInformation("", "", "")
}
