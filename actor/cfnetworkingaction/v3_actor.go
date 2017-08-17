package cfnetworkingaction

import "code.cloudfoundry.org/cli/actor/v3action"

//go:generate counterfeiter . V3Actor
type V3Actor interface {
	GetApplicationByNameAndSpace(appName string, spaceGUID string) (v3action.Application, v3action.Warnings, error)
	GetApplicationsBySpace(spaceGUID string) ([]v3action.Application, v3action.Warnings, error)
}
