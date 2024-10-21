package v7action

import (
	"code.cloudfoundry.org/cli/v7/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/v7/util/configv3"
)

type TargetSettings ccv3.TargetSettings

// SetTarget targets the Cloud Controller using the client and sets target
// information in the actor based on the response.
func (actor Actor) SetTarget(settings TargetSettings) (Warnings, error) {
	rootInfo, warnings, err := actor.CloudControllerClient.TargetCF(ccv3.TargetSettings(settings))
	if err != nil {
		return Warnings(warnings), err
	}

	actor.Config.SetTargetInformation(configv3.TargetInformationArgs{
		Api:               settings.URL,
		ApiVersion:        rootInfo.CloudControllerAPIVersion(),
		Auth:              rootInfo.Login(),
		MinCLIVersion:     "", // Oldest supported V3 version should be OK
		Doppler:           rootInfo.Logging(),
		LogCache:          rootInfo.LogCache(),
		Routing:           rootInfo.Routing(),
		SkipSSLValidation: settings.SkipSSLValidation,
	})
	actor.Config.SetTokenInformation("", "", "")
	return Warnings(warnings), nil
}

// ClearTarget clears target information from the actor.
func (actor Actor) ClearTarget() {
	actor.Config.SetTargetInformation(configv3.TargetInformationArgs{})
	actor.Config.SetTokenInformation("", "", "")
}
