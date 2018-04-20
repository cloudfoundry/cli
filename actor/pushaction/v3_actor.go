package pushaction

import "code.cloudfoundry.org/cli/actor/v3action"

//go:generate counterfeiter . V3Actor

type V3Actor interface {
	CloudControllerAPIVersion() string
	UpdateApplication(v3action.Application) (v3action.Application, v3action.Warnings, error)
}
