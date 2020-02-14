package v7

import (
	"context"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/uaa/constant"
)

//go:generate counterfeiter . Actor

type Actor interface {
	CheckRoute(domainName string, hostname string, path string) (bool, v7action.Warnings, error)
	GetLatestActiveDeploymentForApp(appGUID string) (v7action.Deployment, v7action.Warnings, error)
	CancelDeployment(deploymentGUID string) (v7action.Warnings, error)
	Authenticate(credentials map[string]string, origin string, grantType constant.GrantType) error
	CloudControllerAPIVersion() string
	UAAAPIVersion() string
	GetBuildpacks(labelSelector string) ([]v7action.Buildpack, v7action.Warnings, error)
	GetAppSummariesForSpace(spaceGUID string, labels string) ([]v7action.ApplicationSummary, v7action.Warnings, error)
	SetSpaceManifest(spaceGUID string, rawManifest []byte) (v7action.Warnings, error)
	GetStreamingLogsForApplicationByNameAndSpace(appName string, spaceGUID string, client sharedaction.LogCacheClient) (<-chan sharedaction.LogMessage, <-chan error, context.CancelFunc, v7action.Warnings, error)
	RestartApplication(appGUID string, noWait bool) (v7action.Warnings, error)
	ScaleProcessByApplication(appGUID string, process v7action.Process) (v7action.Warnings, error)
	StopApplication(appGUID string) (v7action.Warnings, error)
	StartApplication(appGUID string) (v7action.Warnings, error)
	PollStart(appGUID string, noWait bool) (v7action.Warnings, error)
	AllowSpaceSSH(spaceName string, orgGUID string) (v7action.Warnings, error)
	GetApplicationByNameAndSpace(name string, spaceGUID string) (v7action.Application, v7action.Warnings, error)
	GetDetailedAppSummary(appName string, spaceGUID string, withObfuscatedValues bool) (v7action.DetailedApplicationSummary, v7action.Warnings, error)
}
