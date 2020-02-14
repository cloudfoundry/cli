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

	var allWarnings Warnings
	warnings, err := actor.CloudControllerClient.TargetCF(ccv3.TargetSettings(settings))
	allWarnings = Warnings(warnings)
	if err != nil {
		return allWarnings, err
	}

	var info ccv3.Info
	info, warnings, err = actor.CloudControllerClient.RootResponse()
	allWarnings = append(allWarnings, Warnings(warnings)...)
	if err != nil {
		return allWarnings, err
	}
	actor.Config.SetTargetInformation(settings.URL,
		info.CloudControllerAPIVersion(),
		info.UAA(),
		"", // Oldest supported V3 version should be OK
		info.Logging(),
		info.Routing(),
		settings.SkipSSLValidation,
	)
	actor.Config.SetTokenInformation("", "", "")
	return allWarnings, nil
}

// ClearTarget clears target information from the actor.
func (actor Actor) ClearTarget() {
	actor.Config.SetTargetInformation("", "", "", "", "", "", false)
	actor.Config.SetTokenInformation("", "", "")
}
