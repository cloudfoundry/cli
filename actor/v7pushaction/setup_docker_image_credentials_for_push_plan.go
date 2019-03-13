package v7pushaction

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/util/manifestparser"
)

func SetupDockerImageCredentialsForPushPlan(pushPlan PushPlan, manifestApp manifestparser.Application) (PushPlan, error) {
	if pushPlan.Application.LifecycleType == constant.AppLifecycleTypeDocker {
		pushPlan.DockerImageCredentialsNeedsUpdate = true

		pushPlan.DockerImageCredentials.Path = pushPlan.Overrides.DockerImage
		pushPlan.DockerImageCredentials.Username = pushPlan.Overrides.DockerUsername
		pushPlan.DockerImageCredentials.Password = pushPlan.Overrides.DockerPassword
	}

	return pushPlan, nil
}
