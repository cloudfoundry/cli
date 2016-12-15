package v3action

import (
	"net/url"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
)

//go:generate counterfeiter . CloudControllerClient

// CloudControllerClient is the interface to the cloud controller V3 API.
type CloudControllerClient interface {
	CloudControllerAPIVersion() string
	GetApplicationTasks(appGUID string, query url.Values) ([]ccv3.Task, ccv3.Warnings, error)
	GetApplications(query url.Values) ([]ccv3.Application, ccv3.Warnings, error)
	NewTask(appGUID string, command string, name string) (ccv3.Task, ccv3.Warnings, error)
	UpdateTask(taskGUID string) (ccv3.Task, ccv3.Warnings, error)
}
