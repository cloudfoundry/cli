package v7pushaction

import (
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/util/manifestparser"
	log "github.com/sirupsen/logrus"
)

func SetupUpdateWebProcessForPushPlan(pushPlan PushPlan, overrides FlagOverrides, manifestApp manifestparser.Application) (PushPlan, error) {
	if shouldUpdateWebProcess(overrides) {
		log.Info("Setting Web Process's Configuration")
		pushPlan.UpdateWebProcessNeedsUpdate = true

		pushPlan.UpdateWebProcess = v7action.Process{
			Command:             overrides.StartCommand,
			HealthCheckType:     overrides.HealthCheckType,
			HealthCheckEndpoint: overrides.HealthCheckEndpoint,
			HealthCheckTimeout:  overrides.HealthCheckTimeout,
		}
	}
	return pushPlan, nil
}

func shouldUpdateWebProcess(overrides FlagOverrides) bool {
	return overrides.StartCommand.IsSet ||
		overrides.HealthCheckType != "" ||
		overrides.HealthCheckTimeout != 0
}
