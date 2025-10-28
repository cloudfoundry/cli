package v7action

import (
	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/v8/util/configv3"
)

type TargetSettings ccv3.TargetSettings

// SetTarget targets the Cloud Controller using the client and sets target
// information in the config based on the response.
func (actor Actor) SetTarget(settings TargetSettings) (Warnings, error) {
	var allWarnings Warnings

	actor.CloudControllerClient.TargetCF(ccv3.TargetSettings(settings))

	rootInfo, warnings, err := actor.CloudControllerClient.GetRoot()
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return allWarnings, err
	}

	actor.Config.SetTargetInformation(configv3.TargetInformationArgs{
		Api:               settings.URL,
		ApiVersion:        rootInfo.CloudControllerAPIVersion(),
		Auth:              rootInfo.Login(),
		MinCLIVersion:     "", // Oldest supported V3 version should be OK
		Doppler:           rootInfo.Logging(),
		LogCache:          rootInfo.LogCache(),
		NetworkPolicyV1:   rootInfo.NetworkPolicyV1(),
		Routing:           rootInfo.Routing(),
		SkipSSLValidation: settings.SkipSSLValidation,
		UAA:               rootInfo.UAA(),
		CFOnK8s:           rootInfo.CFOnK8s,
	})

	actor.Config.SetTokenInformation("", "", "")
	actor.Config.SetKubernetesAuthInfo("")

	return allWarnings, nil
}

// ClearTarget clears target information from the config.
func (actor Actor) ClearTarget() {
	actor.Config.SetTargetInformation(configv3.TargetInformationArgs{})
	actor.Config.SetTokenInformation("", "", "")
	actor.Config.SetKubernetesAuthInfo("")
}
