package v3action

import (
	"net/url"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
)

//go:generate counterfeiter . CloudControllerClient

// CloudControllerClient is the interface to the cloud controller V3 API.
type CloudControllerClient interface {
	CloudControllerAPIVersion() string
	CreateIsolationSegment(name string) (ccv3.IsolationSegment, ccv3.Warnings, error)
	DeleteIsolationSegment(guid string) (ccv3.Warnings, error)
	GetApplications(query url.Values) ([]ccv3.Application, ccv3.Warnings, error)
	GetApplicationTasks(appGUID string, query url.Values) ([]ccv3.Task, ccv3.Warnings, error)
	GetIsolationSegments(query url.Values) ([]ccv3.IsolationSegment, ccv3.Warnings, error)
	NewTask(appGUID string, command string, name string, memory uint64, disk uint64) (ccv3.Task, ccv3.Warnings, error)
	UpdateTask(taskGUID string) (ccv3.Task, ccv3.Warnings, error)
}
