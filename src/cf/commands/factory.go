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

func (f Factory) NewApi() Api {
	return NewApi(
		f.ui,
		f.repoLocator.GetConfigurationRepository(),
	)
}

func (f Factory) NewApps() Apps {
	return NewApps(
		f.ui,
		f.repoLocator.GetSpaceRepository(),
	)
}

func (f Factory) NewBindService() *BindService {
	return NewBindService(
		f.ui,
		f.repoLocator.GetServiceRepository(),
	)
}

func (f Factory) NewCreateService() CreateService {
	return NewCreateService(
		f.ui,
		f.repoLocator.GetServiceRepository(),
	)
}

func (f Factory) NewCreateSpace() CreateSpace {
	return NewCreateSpace(
		f.ui,
		f.repoLocator.GetSpaceRepository(),
	)
}

func (f Factory) NewDelete() *Delete {
	return NewDelete(
		f.ui,
		f.repoLocator.GetApplicationRepository(),
	)
}

func (f Factory) NewDeleteService() *DeleteService {
	return NewDeleteService(
		f.ui,
		f.repoLocator.GetServiceRepository(),
	)
}

func (f Factory) NewDeleteSpace() *DeleteSpace {
	return NewDeleteSpace(
		f.ui,
		f.repoLocator.GetSpaceRepository(),
	)
}

func (f Factory) NewFiles() *Files {
	return NewFiles(
		f.ui,
		f.repoLocator.GetAppFilesRepository(),
	)
}

func (f Factory) NewLogin() Login {
	authenticator := api.NewUAAAuthenticator(f.repoLocator.GetConfigurationRepository())

	return NewLogin(
		f.ui,
		f.repoLocator.GetConfigurationRepository(),
		f.repoLocator.GetOrganizationRepository(),
		f.repoLocator.GetSpaceRepository(),
		authenticator,
	)
}

func (f Factory) NewLogout() Logout {
	return NewLogout(
		f.ui,
		f.repoLocator.GetConfigurationRepository(),
	)
}

func (f Factory) NewMarketplaceServices() MarketplaceServices {
	return NewMarketplaceServices(
		f.ui,
		f.repoLocator.GetServiceRepository(),
	)
}

func (f Factory) NewPassword() Password {
	return NewPassword(
		f.ui,
		f.repoLocator.GetPasswordRepository(),
		f.repoLocator.GetConfigurationRepository(),
	)
}

func (f Factory) NewPush() Push {
	zipper := cf.ApplicationZipper{}

	return NewPush(
		f.ui,
		f.NewStart(),
		zipper,
		f.repoLocator.GetApplicationRepository(),
		f.repoLocator.GetDomainRepository(),
		f.repoLocator.GetRouteRepository(),
		f.repoLocator.GetStackRepository(),
	)
}

func (f Factory) NewRenameOrg() *RenameOrg {
	return NewRenameOrg(
		f.ui,
		f.repoLocator.GetOrganizationRepository(),
	)
}

func (f Factory) NewRenameSpace() *RenameSpace {
	return NewRenameSpace(
		f.ui,
		f.repoLocator.GetSpaceRepository(),
	)
}

func (f Factory) NewRoutes() *Routes {
	return NewRoutes(
		f.ui,
		f.repoLocator.GetRouteRepository(),
	)
}

func (f Factory) NewServices() Services {
	return NewServices(
		f.ui,
		f.repoLocator.GetSpaceRepository(),
	)
}

func (f Factory) NewSetEnv() *SetEnv {
	return NewSetEnv(
		f.ui,
		f.repoLocator.GetApplicationRepository(),
	)
}

func (f Factory) NewSpaces() Spaces {
	return NewSpaces(
		f.ui,
		f.repoLocator.GetConfig(),
		f.repoLocator.GetSpaceRepository(),
	)
}

func (f Factory) NewStacks() *Stacks {
	return NewStacks(
		f.ui,
		f.repoLocator.GetStackRepository(),
	)
}

func (f Factory) NewStart() *Start {
	return NewStart(
		f.ui,
		f.repoLocator.GetConfig(),
		f.repoLocator.GetApplicationRepository(),
	)
}

func (f Factory) NewStop() *Stop {
	return NewStop(
		f.ui,
		f.repoLocator.GetApplicationRepository(),
	)
}

func (f Factory) NewTarget() Target {
	return NewTarget(
		f.ui,
		f.repoLocator.GetConfigurationRepository(),
		f.repoLocator.GetOrganizationRepository(),
		f.repoLocator.GetSpaceRepository(),
	)
}

func (f Factory) NewUnbindService() *UnbindService {
	return NewUnbindService(
		f.ui,
		f.repoLocator.GetServiceRepository(),
	)
}

func (f Factory) NewListOrganizations() ListOrganizations {
	return NewListOrganizations(
		f.ui,
		f.repoLocator.GetOrganizationRepository(),
	)
}

func (f Factory) NewCreateOrganization() CreateOrganization {
	return NewCreateOrganization(
		f.ui,
		f.repoLocator.GetOrganizationRepository(),
	)
}
