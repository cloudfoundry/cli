package requirements

import (
	"cf/api"
	"cf/terminal"
)

type Requirement interface {
	Execute() (success bool)
}

type Factory interface {
	NewApplicationRequirement(name string) ApplicationRequirement
	NewServiceInstanceRequirement(name string) ServiceInstanceRequirement
	NewLoginRequirement() Requirement
	NewSpaceRequirement() Requirement
}

type ApiRequirementFactory struct {
	ui          terminal.UI
	repoLocator api.RepositoryLocator
}

func NewFactory(ui terminal.UI, repoLocator api.RepositoryLocator) (factory ApiRequirementFactory) {
	return ApiRequirementFactory{ui, repoLocator}
}

func (f ApiRequirementFactory) NewApplicationRequirement(name string) ApplicationRequirement {
	return NewApplicationRequirement(
		name,
		f.ui,
		f.repoLocator.GetConfig(),
		f.repoLocator.GetApplicationRepository(),
	)
}

func (f ApiRequirementFactory) NewServiceInstanceRequirement(name string) ServiceInstanceRequirement {
	return NewServiceInstanceRequirement(
		name,
		f.ui,
		f.repoLocator.GetConfig(),
		f.repoLocator.GetServiceRepository(),
	)
}

func (f ApiRequirementFactory) NewLoginRequirement() Requirement {
	return NewLoginRequirement(
		f.ui,
		f.repoLocator.GetConfig(),
	)
}

func (f ApiRequirementFactory) NewSpaceRequirement() Requirement {
	return NewSpaceRequirement(
		f.ui,
		f.repoLocator.GetConfig(),
	)
}
