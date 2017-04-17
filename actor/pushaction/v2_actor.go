package pushaction

import "code.cloudfoundry.org/cli/actor/v2action"

//go:generate counterfeiter . V2Actor

type V2Actor interface {
	CreateApplication(application v2action.Application) (v2action.Application, v2action.Warnings, error)
	GetApplicationByNameAndSpace(name string, spaceGUID string) (v2action.Application, v2action.Warnings, error)
	UpdateApplication(application v2action.Application) (v2action.Application, v2action.Warnings, error)
}
