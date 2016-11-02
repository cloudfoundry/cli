package v3actions

import (
	"net/url"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
)

//go:generate counterfeiter . CloudControllerClient

type CloudControllerClient interface {
	RunTask(appGUID string, command string) (ccv3.Task, ccv3.Warnings, error)
	GetApplications(query url.Values) ([]ccv3.Application, ccv3.Warnings, error)
}
