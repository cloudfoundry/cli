package commands

import (
	"cf/api"
	"cf/commands/application"
	"cf/commands/domain"
	"cf/commands/organization"
	"cf/commands/route"
	"cf/commands/service"
	"cf/commands/space"
	"cf/terminal"
	"errors"
)

type Factory interface {
	GetByCmdName(cmdName string) (cmd Command, err error)
}

type ConcreteFactory struct {
	cmdsByName map[string]Command
}

func NewFactory(ui terminal.UI, repoLocator api.RepositoryLocator) (factory ConcreteFactory) {
	factory.cmdsByName = make(map[string]Command)

	factory.cmdsByName["api"] = NewApi(ui, repoLocator.GetCloudControllerGateway(), repoLocator.GetConfigurationRepository())
	factory.cmdsByName["app"] = application.NewShowApp(ui, repoLocator.GetAppSummaryRepository())
	factory.cmdsByName["apps"] = application.NewListApps(ui, repoLocator.GetSpaceRepository())
	factory.cmdsByName["bind-service"] = service.NewBindService(ui, repoLocator.GetServiceRepository())
	factory.cmdsByName["create-domain"] = domain.NewCreateDomain(ui, repoLocator.GetDomainRepository(), repoLocator.GetOrganizationRepository())
	factory.cmdsByName["create-org"] = organization.NewCreateOrg(ui, repoLocator.GetOrganizationRepository())
	factory.cmdsByName["create-service"] = service.NewCreateService(ui, repoLocator.GetServiceRepository())
	factory.cmdsByName["create-space"] = space.NewCreateSpace(ui, repoLocator.GetSpaceRepository())
	factory.cmdsByName["delete"] = application.NewDeleteApp(ui, repoLocator.GetApplicationRepository())
	factory.cmdsByName["delete-org"] = organization.NewDeleteOrg(ui, repoLocator.GetOrganizationRepository(), repoLocator.GetConfigurationRepository())
	factory.cmdsByName["delete-service"] = service.NewDeleteService(ui, repoLocator.GetServiceRepository())
	factory.cmdsByName["delete-space"] = space.NewDeleteSpace(ui, repoLocator.GetSpaceRepository(), repoLocator.GetConfigurationRepository())
	factory.cmdsByName["env"] = application.NewEnv(ui)
	factory.cmdsByName["files"] = application.NewFiles(ui, repoLocator.GetAppFilesRepository())
	factory.cmdsByName["login"] = NewLogin(ui, repoLocator.GetConfigurationRepository(), repoLocator.GetAuthenticator())
	factory.cmdsByName["logout"] = NewLogout(ui, repoLocator.GetConfigurationRepository())
	factory.cmdsByName["logs"] = application.NewLogs(ui, repoLocator.GetLogsRepository())
	factory.cmdsByName["marketplace"] = service.NewMarketplaceServices(ui, repoLocator.GetServiceRepository())
	factory.cmdsByName["org"] = organization.NewShowOrg(ui)
	factory.cmdsByName["orgs"] = organization.NewListOrgs(ui, repoLocator.GetOrganizationRepository())
	factory.cmdsByName["password"] = NewPassword(ui, repoLocator.GetPasswordRepository(), repoLocator.GetConfigurationRepository())
	factory.cmdsByName["rename"] = application.NewRenameApp(ui, repoLocator.GetApplicationRepository())
	factory.cmdsByName["rename-org"] = organization.NewRenameOrg(ui, repoLocator.GetOrganizationRepository())
	factory.cmdsByName["rename-service"] = service.NewRenameService(ui, repoLocator.GetServiceRepository())
	factory.cmdsByName["rename-space"] = space.NewRenameSpace(ui, repoLocator.GetSpaceRepository(), repoLocator.GetConfigurationRepository())
	factory.cmdsByName["routes"] = route.NewListRoutes(ui, repoLocator.GetRouteRepository())
	factory.cmdsByName["set-env"] = application.NewSetEnv(ui, repoLocator.GetApplicationRepository())
	factory.cmdsByName["space"] = space.NewShowSpace(ui, repoLocator.GetConfig())
	factory.cmdsByName["service"] = service.NewShowService(ui, repoLocator.GetServiceRepository())
	factory.cmdsByName["services"] = service.NewListServices(ui, repoLocator.GetSpaceRepository())
	factory.cmdsByName["spaces"] = space.NewListSpaces(ui, repoLocator.GetConfig(), repoLocator.GetSpaceRepository())
	factory.cmdsByName["stacks"] = NewStacks(ui, repoLocator.GetStackRepository())
	factory.cmdsByName["target"] = NewTarget(ui, repoLocator.GetConfigurationRepository(), repoLocator.GetOrganizationRepository(), repoLocator.GetSpaceRepository())
	factory.cmdsByName["unbind-service"] = service.NewUnbindService(ui, repoLocator.GetServiceRepository())
	factory.cmdsByName["unset-env"] = application.NewUnsetEnv(ui, repoLocator.GetApplicationRepository())

	start := application.NewStart(ui, repoLocator.GetConfig(), repoLocator.GetApplicationRepository())
	stop := application.NewStop(ui, repoLocator.GetApplicationRepository())
	restart := application.NewRestart(ui, start, stop)

	factory.cmdsByName["start"] = start
	factory.cmdsByName["stop"] = stop
	factory.cmdsByName["restart"] = restart
	factory.cmdsByName["push"] = application.NewPush(ui, start, stop, repoLocator.GetApplicationRepository(), repoLocator.GetDomainRepository(), repoLocator.GetRouteRepository(), repoLocator.GetStackRepository(), repoLocator.GetApplicationBitsRepository())
	factory.cmdsByName["scale"] = application.NewScale(ui, restart, repoLocator.GetApplicationRepository())

	return
}

func (f ConcreteFactory) GetByCmdName(cmdName string) (cmd Command, err error) {
	cmd, found := f.cmdsByName[cmdName]
	if !found {
		err = errors.New("Command not found")
	}
	return
}
