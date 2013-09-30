package commands

import (
	"cf/api"
	"cf/commands/organization"
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
	factory.cmdsByName["app"] = NewApp(ui, repoLocator.GetAppSummaryRepository())
	factory.cmdsByName["apps"] = NewApps(ui, repoLocator.GetSpaceRepository())
	factory.cmdsByName["bind-service"] = NewBindService(ui, repoLocator.GetServiceRepository())
	factory.cmdsByName["create-domain"] = NewCreateDomain(ui, repoLocator.GetDomainRepository(), repoLocator.GetOrganizationRepository())
	factory.cmdsByName["create-org"] = organization.NewCreateOrg(ui, repoLocator.GetOrganizationRepository())
	factory.cmdsByName["create-service"] = NewCreateService(ui, repoLocator.GetServiceRepository())
	factory.cmdsByName["create-space"] = NewCreateSpace(ui, repoLocator.GetSpaceRepository())
	factory.cmdsByName["delete"] = NewDelete(ui, repoLocator.GetApplicationRepository())
	factory.cmdsByName["delete-org"] = organization.NewDeleteOrg(ui, repoLocator.GetOrganizationRepository(), repoLocator.GetConfigurationRepository())
	factory.cmdsByName["delete-service"] = NewDeleteService(ui, repoLocator.GetServiceRepository())
	factory.cmdsByName["delete-space"] = NewDeleteSpace(ui, repoLocator.GetSpaceRepository())
	factory.cmdsByName["env"] = NewEnv(ui)
	factory.cmdsByName["files"] = NewFiles(ui, repoLocator.GetAppFilesRepository())
	factory.cmdsByName["login"] = NewLogin(ui, repoLocator.GetConfigurationRepository(), repoLocator.GetAuthenticator())
	factory.cmdsByName["logout"] = NewLogout(ui, repoLocator.GetConfigurationRepository())
	factory.cmdsByName["logs"] = NewLogs(ui, repoLocator.GetLogsRepository())
	factory.cmdsByName["logs-recent"] = NewRecentLogs(ui, repoLocator.GetApplicationRepository(), repoLocator.GetLogsRepository())
	factory.cmdsByName["marketplace"] = NewMarketplaceServices(ui, repoLocator.GetServiceRepository())
	factory.cmdsByName["org"] = organization.NewShowOrg(ui)
	factory.cmdsByName["orgs"] = organization.NewListOrgs(ui, repoLocator.GetOrganizationRepository())
	factory.cmdsByName["password"] = NewPassword(ui, repoLocator.GetPasswordRepository(), repoLocator.GetConfigurationRepository())
	factory.cmdsByName["rename"] = NewRename(ui, repoLocator.GetApplicationRepository())
	factory.cmdsByName["rename-org"] = organization.NewRenameOrg(ui, repoLocator.GetOrganizationRepository())
	factory.cmdsByName["rename-service"] = NewRenameService(ui, repoLocator.GetServiceRepository())
	factory.cmdsByName["rename-space"] = NewRenameSpace(ui, repoLocator.GetSpaceRepository(), repoLocator.GetConfigurationRepository())
	factory.cmdsByName["routes"] = NewRoutes(ui, repoLocator.GetRouteRepository())
	factory.cmdsByName["set-env"] = NewSetEnv(ui, repoLocator.GetApplicationRepository())
	factory.cmdsByName["space"] = NewShowSpace(ui, repoLocator.GetConfig())
	factory.cmdsByName["service"] = NewShowService(ui, repoLocator.GetServiceRepository())
	factory.cmdsByName["services"] = NewServices(ui, repoLocator.GetSpaceRepository())
	factory.cmdsByName["spaces"] = NewSpaces(ui, repoLocator.GetConfig(), repoLocator.GetSpaceRepository())
	factory.cmdsByName["stacks"] = NewStacks(ui, repoLocator.GetStackRepository())
	factory.cmdsByName["target"] = NewTarget(ui, repoLocator.GetConfigurationRepository(), repoLocator.GetOrganizationRepository(), repoLocator.GetSpaceRepository())
	factory.cmdsByName["unbind-service"] = NewUnbindService(ui, repoLocator.GetServiceRepository())
	factory.cmdsByName["unset-env"] = NewUnsetEnv(ui, repoLocator.GetApplicationRepository())

	start := NewStart(ui, repoLocator.GetConfig(), repoLocator.GetApplicationRepository())
	stop := NewStop(ui, repoLocator.GetApplicationRepository())
	restart := NewRestart(ui, start, stop)

	factory.cmdsByName["start"] = start
	factory.cmdsByName["stop"] = stop
	factory.cmdsByName["restart"] = restart
	factory.cmdsByName["push"] = NewPush(ui, start, stop, repoLocator.GetApplicationRepository(), repoLocator.GetDomainRepository(), repoLocator.GetRouteRepository(), repoLocator.GetStackRepository(), repoLocator.GetApplicationBitsRepository())
	factory.cmdsByName["scale"] = NewScale(ui, restart, repoLocator.GetApplicationRepository())

	return
}

func (f ConcreteFactory) GetByCmdName(cmdName string) (cmd Command, err error) {
	cmd, found := f.cmdsByName[cmdName]
	if !found {
		err = errors.New("Command not found")
	}
	return
}
