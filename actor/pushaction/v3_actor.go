package pushaction

import "code.cloudfoundry.org/cli/actor/v3action"

//go:generate counterfeiter . V3Actor

type V3Actor interface {
	CloudControllerAPIVersion() string
	CreateApplicationInSpace(app v3action.Application, spaceGUID string) (v3action.Application, v3action.Warnings, error)
	GetApplicationByNameAndSpace(appName string, spaceGUID string) (v3action.Application, v3action.Warnings, error)
	UpdateApplication(v3action.Application) (v3action.Application, v3action.Warnings, error)
}
