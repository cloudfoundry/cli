package commands

import (
	"cf/api"
	"cf/commands/application"
	"cf/commands/buildpack"
	"cf/commands/domain"
	"cf/commands/organization"
	"cf/commands/route"
	"cf/commands/service"
	"cf/commands/serviceauthtoken"
	"cf/commands/servicebroker"
	"cf/commands/space"
	"cf/commands/user"
	"cf/configuration"
	"cf/manifest"
	"cf/terminal"
	"errors"
)

type Factory interface {
	GetByCmdName(cmdName string) (cmd Command, err error)
}

type ConcreteFactory struct {
	cmdsByName map[string]Command
}

func NewFactory(ui terminal.UI, config configuration.ReadWriter, manifestRepo manifest.ManifestRepository, repoLocator api.RepositoryLocator) (factory ConcreteFactory) {
	factory.cmdsByName = make(map[string]Command)

	factory.cmdsByName["api"] = NewApi(ui, config, repoLocator.GetEndpointRepository())
	factory.cmdsByName["apps"] = application.NewListApps(ui, config, repoLocator.GetAppSummaryRepository())
	factory.cmdsByName["auth"] = NewAuthenticate(ui, config, repoLocator.GetAuthenticationRepository())
	factory.cmdsByName["buildpacks"] = buildpack.NewListBuildpacks(ui, repoLocator.GetBuildpackRepository())
	factory.cmdsByName["create-buildpack"] = buildpack.NewCreateBuildpack(ui, repoLocator.GetBuildpackRepository(), repoLocator.GetBuildpackBitsRepository())
	factory.cmdsByName["create-domain"] = domain.NewCreateDomain(ui, config, repoLocator.GetDomainRepository())
	factory.cmdsByName["create-org"] = organization.NewCreateOrg(ui, config, repoLocator.GetOrganizationRepository())
	factory.cmdsByName["create-service"] = service.NewCreateService(ui, config, repoLocator.GetServiceRepository())
	factory.cmdsByName["create-service-auth-token"] = serviceauthtoken.NewCreateServiceAuthToken(ui, config, repoLocator.GetServiceAuthTokenRepository())
	factory.cmdsByName["create-service-broker"] = servicebroker.NewCreateServiceBroker(ui, config, repoLocator.GetServiceBrokerRepository())
	factory.cmdsByName["create-user"] = user.NewCreateUser(ui, config, repoLocator.GetUserRepository())
	factory.cmdsByName["create-user-provided-service"] = service.NewCreateUserProvidedService(ui, config, repoLocator.GetUserProvidedServiceInstanceRepository())
	factory.cmdsByName["curl"] = NewCurl(ui, config, repoLocator.GetCurlRepository())
	factory.cmdsByName["delete"] = application.NewDeleteApp(ui, config, repoLocator.GetApplicationRepository())
	factory.cmdsByName["delete-buildpack"] = buildpack.NewDeleteBuildpack(ui, repoLocator.GetBuildpackRepository())
	factory.cmdsByName["delete-domain"] = domain.NewDeleteDomain(ui, config, repoLocator.GetDomainRepository())
	factory.cmdsByName["delete-shared-domain"] = domain.NewDeleteSharedDomain(ui, config, repoLocator.GetDomainRepository())
	factory.cmdsByName["delete-org"] = organization.NewDeleteOrg(ui, config, repoLocator.GetOrganizationRepository())
	factory.cmdsByName["delete-route"] = route.NewDeleteRoute(ui, config, repoLocator.GetRouteRepository())
	factory.cmdsByName["delete-service"] = service.NewDeleteService(ui, config, repoLocator.GetServiceRepository())
	factory.cmdsByName["delete-service-auth-token"] = serviceauthtoken.NewDeleteServiceAuthToken(ui, config, repoLocator.GetServiceAuthTokenRepository())
	factory.cmdsByName["delete-service-broker"] = servicebroker.NewDeleteServiceBroker(ui, config, repoLocator.GetServiceBrokerRepository())
	factory.cmdsByName["delete-space"] = space.NewDeleteSpace(ui, config, repoLocator.GetSpaceRepository())
	factory.cmdsByName["delete-user"] = user.NewDeleteUser(ui, config, repoLocator.GetUserRepository())
	factory.cmdsByName["domains"] = domain.NewListDomains(ui, config, repoLocator.GetDomainRepository())
	factory.cmdsByName["env"] = application.NewEnv(ui, config)
	factory.cmdsByName["events"] = application.NewEvents(ui, config, repoLocator.GetAppEventsRepository())
	factory.cmdsByName["files"] = application.NewFiles(ui, config, repoLocator.GetAppFilesRepository())
	factory.cmdsByName["login"] = NewLogin(ui, config, repoLocator.GetAuthenticationRepository(), repoLocator.GetEndpointRepository(), repoLocator.GetOrganizationRepository(), repoLocator.GetSpaceRepository())
	factory.cmdsByName["logout"] = NewLogout(ui, config)
	factory.cmdsByName["logs"] = application.NewLogs(ui, config, repoLocator.GetLogsRepository())
	factory.cmdsByName["marketplace"] = service.NewMarketplaceServices(ui, config, repoLocator.GetServiceRepository())
	factory.cmdsByName["org"] = organization.NewShowOrg(ui, config)
	factory.cmdsByName["org-users"] = user.NewOrgUsers(ui, config, repoLocator.GetUserRepository())
	factory.cmdsByName["orgs"] = organization.NewListOrgs(ui, config, repoLocator.GetOrganizationRepository())
	factory.cmdsByName["passwd"] = NewPassword(ui, repoLocator.GetPasswordRepository(), config)
	factory.cmdsByName["purge-service-offering"] = service.NewPurgeServiceOffering(ui, config, repoLocator.GetServiceRepository())
	factory.cmdsByName["quotas"] = organization.NewListQuotas(ui, config, repoLocator.GetQuotaRepository())
	factory.cmdsByName["rename"] = application.NewRenameApp(ui, config, repoLocator.GetApplicationRepository())
	factory.cmdsByName["rename-org"] = organization.NewRenameOrg(ui, config, repoLocator.GetOrganizationRepository())
	factory.cmdsByName["rename-service"] = service.NewRenameService(ui, config, repoLocator.GetServiceRepository())
	factory.cmdsByName["rename-service-broker"] = servicebroker.NewRenameServiceBroker(ui, config, repoLocator.GetServiceBrokerRepository())
	factory.cmdsByName["rename-space"] = space.NewRenameSpace(ui, config, repoLocator.GetSpaceRepository())
	factory.cmdsByName["routes"] = route.NewListRoutes(ui, config, repoLocator.GetRouteRepository())
	factory.cmdsByName["service"] = service.NewShowService(ui)
	factory.cmdsByName["service-auth-tokens"] = serviceauthtoken.NewListServiceAuthTokens(ui, config, repoLocator.GetServiceAuthTokenRepository())
	factory.cmdsByName["service-brokers"] = servicebroker.NewListServiceBrokers(ui, config, repoLocator.GetServiceBrokerRepository())
	factory.cmdsByName["services"] = service.NewListServices(ui, config, repoLocator.GetServiceSummaryRepository())
	factory.cmdsByName["migrate-service-instances"] = service.NewMigrateServiceInstances(ui, config, repoLocator.GetServiceRepository())
	factory.cmdsByName["set-env"] = application.NewSetEnv(ui, config, repoLocator.GetApplicationRepository())
	factory.cmdsByName["set-org-role"] = user.NewSetOrgRole(ui, config, repoLocator.GetUserRepository())
	factory.cmdsByName["set-quota"] = organization.NewSetQuota(ui, config, repoLocator.GetQuotaRepository())
	factory.cmdsByName["create-shared-domain"] = domain.NewCreateSharedDomain(ui, config, repoLocator.GetDomainRepository())
	factory.cmdsByName["space"] = space.NewShowSpace(ui, config)
	factory.cmdsByName["space-users"] = user.NewSpaceUsers(ui, config, repoLocator.GetSpaceRepository(), repoLocator.GetUserRepository())
	factory.cmdsByName["spaces"] = space.NewListSpaces(ui, config, repoLocator.GetSpaceRepository())
	factory.cmdsByName["stacks"] = NewStacks(ui, config, repoLocator.GetStackRepository())
	factory.cmdsByName["target"] = NewTarget(ui, config, repoLocator.GetOrganizationRepository(), repoLocator.GetSpaceRepository())
	factory.cmdsByName["unbind-service"] = service.NewUnbindService(ui, config, repoLocator.GetServiceBindingRepository())
	factory.cmdsByName["unset-env"] = application.NewUnsetEnv(ui, config, repoLocator.GetApplicationRepository())
	factory.cmdsByName["unset-org-role"] = user.NewUnsetOrgRole(ui, config, repoLocator.GetUserRepository())
	factory.cmdsByName["unset-space-role"] = user.NewUnsetSpaceRole(ui, config, repoLocator.GetSpaceRepository(), repoLocator.GetUserRepository())
	factory.cmdsByName["update-buildpack"] = buildpack.NewUpdateBuildpack(ui, repoLocator.GetBuildpackRepository(), repoLocator.GetBuildpackBitsRepository())
	factory.cmdsByName["update-service-broker"] = servicebroker.NewUpdateServiceBroker(ui, config, repoLocator.GetServiceBrokerRepository())
	factory.cmdsByName["update-service-auth-token"] = serviceauthtoken.NewUpdateServiceAuthToken(ui, config, repoLocator.GetServiceAuthTokenRepository())
	factory.cmdsByName["update-user-provided-service"] = service.NewUpdateUserProvidedService(ui, config, repoLocator.GetUserProvidedServiceInstanceRepository())

	createRoute := route.NewCreateRoute(ui, config, repoLocator.GetRouteRepository())
	factory.cmdsByName["create-route"] = createRoute
	factory.cmdsByName["map-route"] = route.NewMapRoute(ui, config, repoLocator.GetRouteRepository(), createRoute)
	factory.cmdsByName["unmap-route"] = route.NewUnmapRoute(ui, config, repoLocator.GetRouteRepository())

	displayApp := application.NewShowApp(ui, config, repoLocator.GetAppSummaryRepository(), repoLocator.GetAppInstancesRepository())
	start := application.NewStart(ui, config, displayApp, repoLocator.GetApplicationRepository(), repoLocator.GetAppInstancesRepository(), repoLocator.GetLogsRepository())
	stop := application.NewStop(ui, config, repoLocator.GetApplicationRepository())
	restart := application.NewRestart(ui, start, stop)
	bind := service.NewBindService(ui, config, repoLocator.GetServiceBindingRepository())

	factory.cmdsByName["app"] = displayApp
	factory.cmdsByName["bind-service"] = bind
	factory.cmdsByName["start"] = start
	factory.cmdsByName["stop"] = stop
	factory.cmdsByName["restart"] = restart
	factory.cmdsByName["push"] = application.NewPush(ui, config, manifestRepo, start, stop, bind, repoLocator.GetApplicationRepository(), repoLocator.GetDomainRepository(), repoLocator.GetRouteRepository(), repoLocator.GetStackRepository(), repoLocator.GetServiceRepository(), repoLocator.GetApplicationBitsRepository())
	factory.cmdsByName["scale"] = application.NewScale(ui, config, restart, repoLocator.GetApplicationRepository())

	spaceRoleSetter := user.NewSetSpaceRole(ui, config, repoLocator.GetSpaceRepository(), repoLocator.GetUserRepository())
	factory.cmdsByName["set-space-role"] = spaceRoleSetter
	factory.cmdsByName["create-space"] = space.NewCreateSpace(ui, config, spaceRoleSetter, repoLocator.GetSpaceRepository(), repoLocator.GetOrganizationRepository(), repoLocator.GetUserRepository())

	return
}

func (f ConcreteFactory) GetByCmdName(cmdName string) (cmd Command, err error) {
	cmd, found := f.cmdsByName[cmdName]
	if !found {
		err = errors.New("Command not found")
	}
	return
}
