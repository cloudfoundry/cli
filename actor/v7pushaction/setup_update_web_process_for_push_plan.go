package v7pushaction

import (
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/util/manifestparser"
	log "github.com/sirupsen/logrus"
)

func SetupUpdateWebProcessForPushPlan(pushPlan PushPlan, manifestApp manifestparser.Application) (PushPlan, error) {
	if shouldUpdateWebProcess(pushPlan) {
		log.Info("Setting Web Process's Configuration")
		pushPlan.UpdateWebProcessNeedsUpdate = true

		pushPlan.UpdateWebProcess = v7action.Process{
			Command:             pushPlan.Overrides.StartCommand,
			HealthCheckType:     pushPlan.Overrides.HealthCheckType,
			HealthCheckEndpoint: pushPlan.Overrides.HealthCheckEndpoint,
			HealthCheckTimeout:  pushPlan.Overrides.HealthCheckTimeout,
		}
	}
	return pushPlan, nil
}

func shouldUpdateWebProcess(pushPlan PushPlan) bool {
	return pushPlan.Overrides.StartCommand.IsSet ||
		pushPlan.Overrides.HealthCheckType != "" ||
		pushPlan.Overrides.HealthCheckTimeout != 0
}
