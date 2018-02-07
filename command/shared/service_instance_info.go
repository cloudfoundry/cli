package shared

import "code.cloudfoundry.org/cli/command"

func DisplayServiceInstanceNotShareable(ui command.UI, featureFlagEnabled bool, serviceBrokerSharingEnabled bool) {
	switch {
	case !featureFlagEnabled && !serviceBrokerSharingEnabled:
		ui.DisplayNewline()
		ui.DisplayText(`The "service_instance_sharing" feature flag is disabled for this Cloud Foundry platform. Also, service instance sharing is disabled for this service.`)
	case !featureFlagEnabled && serviceBrokerSharingEnabled:
		ui.DisplayNewline()
		ui.DisplayText(`The "service_instance_sharing" feature flag is disabled for this Cloud Foundry platform.`)
	case featureFlagEnabled && !serviceBrokerSharingEnabled:
		ui.DisplayNewline()
		ui.DisplayText("Service instance sharing is disabled for this service.")
	}
}
