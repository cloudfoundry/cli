package shared

import "code.cloudfoundry.org/cli/actor/v2v3action"

//go:generate counterfeiter . ApplicationSummaryActor

type ApplicationSummaryActor interface {
	CloudControllerV3APIVersion() string
	GetApplicationSummaryByNameAndSpace(appName string, spaceGUID string, withObfuscatedValues bool) (v2v3action.ApplicationSummary, v2v3action.Warnings, error)
}
