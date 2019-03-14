package v7pushaction

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/util/manifestparser"
)

func SetupDockerImageCredentialsForPushPlan(pushPlan PushPlan, overrides FlagOverrides, manifestApp manifestparser.Application) (PushPlan, error) {
	if pushPlan.Application.LifecycleType == constant.AppLifecycleTypeDocker {
		pushPlan.DockerImageCredentialsNeedsUpdate = true

		pushPlan.DockerImageCredentials.Path = overrides.DockerImage
		pushPlan.DockerImageCredentials.Username = overrides.DockerUsername
		pushPlan.DockerImageCredentials.Password = overrides.DockerPassword
	}

	return pushPlan, nil
}
