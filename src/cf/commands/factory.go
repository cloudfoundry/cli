package commands

import (
	"cf"
	"cf/api"
	"cf/terminal"
)

type Factory struct {
	ui          terminal.UI
	repoLocator api.RepositoryLocator
}

func NewFactory(ui terminal.UI, repoLocator api.RepositoryLocator) (factory Factory) {
	factory.ui = ui
	factory.repoLocator = repoLocator
	return
}

func (f Factory) NewTarget() Target {
	return NewTarget(
		f.ui,
		f.repoLocator.GetConfig(),
		f.repoLocator.GetOrganizationRepository(),
		f.repoLocator.GetSpaceRepository(),
	)
}

func (f Factory) NewLogin() Login {
	authenticator := new(api.UAAAuthenticator)

	return NewLogin(
		f.ui,
		f.repoLocator.GetConfig(),
		f.repoLocator.GetOrganizationRepository(),
		f.repoLocator.GetSpaceRepository(),
		authenticator,
	)
}

func (f Factory) NewSetEnv() *SetEnv {
	return NewSetEnv(
		f.ui,
		f.repoLocator.GetConfig(),
		f.repoLocator.GetApplicationRepository(),
	)
}

func (f Factory) NewLogout() Logout {
	return NewLogout(
		f.ui,
		f.repoLocator.GetConfig(),
	)
}

func (f Factory) NewStart() *Start {
	return NewStart(
		f.ui,
		f.repoLocator.GetConfig(),
		f.repoLocator.GetApplicationRepository(),
	)
}

func (f Factory) NewPush() Push {
	zipper := cf.ApplicationZipper{}

	return NewPush(
		f.ui,
		f.repoLocator.GetConfig(),
		f.NewStart(),
		zipper,
		f.repoLocator.GetApplicationRepository(),
		f.repoLocator.GetDomainRepository(),
		f.repoLocator.GetRouteRepository(),
		f.repoLocator.GetStackRepository(),
	)
}

func (f Factory) NewApps() Apps {
	return NewApps(
		f.ui,
		f.repoLocator.GetConfig(),
		f.repoLocator.GetSpaceRepository(),
	)
}

func (f Factory) NewDelete() *Delete {
	return NewDelete(
		f.ui,
		f.repoLocator.GetConfig(),
		f.repoLocator.GetApplicationRepository(),
	)
}

func (f Factory) NewStop() Stop {
	return NewStop(
		f.ui,
		f.repoLocator.GetConfig(),
		f.repoLocator.GetApplicationRepository(),
	)
}

func (f Factory) NewCreateService() CreateService {
	return NewCreateService(
		f.ui,
		f.repoLocator.GetConfig(),
		f.repoLocator.GetServiceRepository(),
	)
}

func (f Factory) NewBindService() *BindService {
	return NewBindService(
		f.ui,
		f.repoLocator.GetConfig(),
		f.repoLocator.GetServiceRepository(),
	)
}

func (f Factory) NewUnbindService() *UnbindService {
	return NewUnbindService(
		f.ui,
		f.repoLocator.GetConfig(),
		f.repoLocator.GetServiceRepository(),
	)
}

func (f Factory) NewDeleteService() *DeleteService {
	return NewDeleteService(
		f.ui,
		f.repoLocator.GetConfig(),
		f.repoLocator.GetServiceRepository(),
	)
}
