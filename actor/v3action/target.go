package v3action

import (
	"strings"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/util/configv3"
)

type TargetSettings ccv3.TargetSettings

// SetTarget targets the Cloud Controller using the client and sets target
// information in the actor based on the response.
func (actor Actor) SetTarget(settings TargetSettings) (Warnings, error) {
	if actor.Config.Target() == settings.URL && actor.Config.SkipSSLValidation() == settings.SkipSSLValidation {
		return nil, nil
	}

	info, warnings, err := actor.CloudControllerClient.TargetCF(ccv3.TargetSettings(settings))
	if err != nil {
		return Warnings(warnings), err
	}
	//TODO Remove this condition when earliest supportest CAPI is 1.87.0
	//We have to do this because the current legacy supported CAPI version as of 2020 does not display the log cache url, this will break if a foundation on legacy CAPI have non-standard logcache urls

	logCacheUrl := info.LogCache()
	if logCacheUrl == "" {
		logCacheUrl = strings.Replace(settings.URL, "api", "log-cache", 1)
	}
	actor.Config.SetTargetInformation(configv3.TargetInformationArgs{
		Api:               settings.URL,
		ApiVersion:        info.CloudControllerAPIVersion(),
		Auth:              info.UAA(),
		MinCLIVersion:     "", // Oldest supported V3 version should be OK
		Doppler:           info.Logging(),
		LogCache:          logCacheUrl,
		Routing:           info.Routing(),
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
