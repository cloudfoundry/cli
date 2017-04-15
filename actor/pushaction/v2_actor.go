package pushaction

import "code.cloudfoundry.org/cli/actor/v2action"

//go:generate counterfeiter . V2Actor

type V2Actor interface {
	GetApplicationByNameAndSpace(name string, spaceGUID string) (v2action.Application, v2action.Warnings, error)
}
