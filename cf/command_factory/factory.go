package command_factory

import (
	"errors"

	actor_plugin_repo "github.com/cloudfoundry/cli/cf/actors/plugin_repo"
	"github.com/cloudfoundry/cli/plugin/rpc"
	"github.com/cloudfoundry/cli/utils"

	"github.com/cloudfoundry/cli/cf/actors/plan_builder"
	"github.com/cloudfoundry/cli/cf/actors/service_builder"
	"github.com/cloudfoundry/cli/cf/flag_helpers"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/codegangsta/cli"

	"github.com/cloudfoundry/cli/cf/actors"
	"github.com/cloudfoundry/cli/cf/actors/broker_builder"
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/app_files"
	"github.com/cloudfoundry/cli/cf/command"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/commands"
	"github.com/cloudfoundry/cli/cf/commands/application"
	"github.com/cloudfoundry/cli/cf/commands/buildpack"
	"github.com/cloudfoundry/cli/cf/commands/domain"
	"github.com/cloudfoundry/cli/cf/commands/environmentvariablegroup"
	"github.com/cloudfoundry/cli/cf/commands/featureflag"
	"github.com/cloudfoundry/cli/cf/commands/organization"
	"github.com/cloudfoundry/cli/cf/commands/plugin"
	"github.com/cloudfoundry/cli/cf/commands/plugin_repo"
	"github.com/cloudfoundry/cli/cf/commands/quota"
	"github.com/cloudfoundry/cli/cf/commands/route"
	"github.com/cloudfoundry/cli/cf/commands/securitygroup"
	"github.com/cloudfoundry/cli/cf/commands/service"
	"github.com/cloudfoundry/cli/cf/commands/serviceaccess"
	"github.com/cloudfoundry/cli/cf/commands/serviceauthtoken"
	"github.com/cloudfoundry/cli/cf/commands/servicebroker"
	"github.com/cloudfoundry/cli/cf/commands/servicekey"
	"github.com/cloudfoundry/cli/cf/commands/space"
	"github.com/cloudfoundry/cli/cf/commands/spacequota"
	"github.com/cloudfoundry/cli/cf/commands/user"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/configuration/plugin_config"
	"github.com/cloudfoundry/cli/cf/manifest"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/words/generator"
)

type Factory interface {
	GetByCmdName(cmdName string) (cmd command.Command, err error)
	CommandMetadatas() []command_metadata.CommandMetadata
	CheckIfCoreCmdExists(cmdName string) bool
	GetCommandFlags(string) []string
	GetCommandTotalArgs(string) (int, error)
}

type concreteFactory struct {
	cmdsByName map[string]command.Command
}

func NewFactory(ui terminal.UI, config core_config.ReadWriter, manifestRepo manifest.ManifestRepository, repoLocator api.RepositoryLocator, pluginConfig plugin_config.PluginConfiguration, rpcService *rpc.CliRpcService) (factory concreteFactory) {
	factory.cmdsByName = make(map[string]command.Command)

	planBuilder := plan_builder.NewBuilder(
		repoLocator.GetServicePlanRepository(),
		repoLocator.GetServicePlanVisibilityRepository(),
		repoLocator.GetOrganizationRepository(),
	)

	serviceBuilder := service_builder.NewBuilder(
		repoLocator.GetServiceRepository(),
		planBuilder,
	)

	brokerBuilder := broker_builder.NewBuilder(
		repoLocator.GetServiceBrokerRepository(),
		serviceBuilder,
	)

	factory.cmdsByName["api"] = commands.NewApi(ui, config, repoLocator.GetEndpointRepository())
	factory.cmdsByName["apps"] = application.NewListApps(ui, config, repoLocator.GetAppSummaryRepository(), repoLocator.GetSpaceRepository())
	factory.cmdsByName["auth"] = commands.NewAuthenticate(ui, config, repoLocator.GetAuthenticationRepository())
	factory.cmdsByName["buildpacks"] = buildpack.NewListBuildpacks(ui, repoLocator.GetBuildpackRepository())
	factory.cmdsByName["config"] = commands.NewConfig(ui, config)
	factory.cmdsByName["create-app-manifest"] = commands.NewCreateAppManifest(ui, config, repoLocator.GetAppSummaryRepository(), manifest.NewGenerator())
	factory.cmdsByName["create-buildpack"] = buildpack.NewCreateBuildpack(ui, repoLocator.GetBuildpackRepository(), repoLocator.GetBuildpackBitsRepository())
	factory.cmdsByName["create-domain"] = domain.NewCreateDomain(ui, config, repoLocator.GetDomainRepository())
	factory.cmdsByName["create-org"] = organization.NewCreateOrg(ui, config, repoLocator.GetOrganizationRepository(), repoLocator.GetQuotaRepository())
	factory.cmdsByName["create-service"] = service.NewCreateService(ui, config, repoLocator.GetServiceRepository(), serviceBuilder)

	factory.cmdsByName["update-service"] = service.NewUpdateService(
		ui,
		config,
		repoLocator.GetServiceRepository(),
		plan_builder.NewBuilder(
			repoLocator.GetServicePlanRepository(),
			repoLocator.GetServicePlanVisibilityRepository(),
			repoLocator.GetOrganizationRepository(),
		),
	)

	factory.cmdsByName["create-service-auth-token"] = serviceauthtoken.NewCreateServiceAuthToken(ui, config, repoLocator.GetServiceAuthTokenRepository())
	factory.cmdsByName["create-service-broker"] = servicebroker.NewCreateServiceBroker(ui, config, repoLocator.GetServiceBrokerRepository())
	factory.cmdsByName["create-service-key"] = servicekey.NewCreateServiceKey(ui, config, repoLocator.GetServiceRepository(), repoLocator.GetServiceKeyRepository())
	factory.cmdsByName["create-user"] = user.NewCreateUser(ui, config, repoLocator.GetUserRepository())
	factory.cmdsByName["create-user-provided-service"] = service.NewCreateUserProvidedService(ui, config, repoLocator.GetUserProvidedServiceInstanceRepository())
	factory.cmdsByName["curl"] = commands.NewCurl(ui, config, repoLocator.GetCurlRepository())
	factory.cmdsByName["delete"] = application.NewDeleteApp(ui, config, repoLocator.GetApplicationRepository(), repoLocator.GetRouteRepository())
	factory.cmdsByName["delete-buildpack"] = buildpack.NewDeleteBuildpack(ui, repoLocator.GetBuildpackRepository())
	factory.cmdsByName["delete-domain"] = domain.NewDeleteDomain(ui, config, repoLocator.GetDomainRepository())
	factory.cmdsByName["delete-shared-domain"] = domain.NewDeleteSharedDomain(ui, config, repoLocator.GetDomainRepository())
	factory.cmdsByName["delete-org"] = organization.NewDeleteOrg(ui, config, repoLocator.GetOrganizationRepository())
	factory.cmdsByName["delete-orphaned-routes"] = route.NewDeleteOrphanedRoutes(ui, config, repoLocator.GetRouteRepository())
	factory.cmdsByName["delete-route"] = route.NewDeleteRoute(ui, config, repoLocator.GetRouteRepository())
	factory.cmdsByName["delete-service"] = service.NewDeleteService(ui, config, repoLocator.GetServiceRepository())
	factory.cmdsByName["delete-service-auth-token"] = serviceauthtoken.NewDeleteServiceAuthToken(ui, config, repoLocator.GetServiceAuthTokenRepository())
	factory.cmdsByName["delete-service-broker"] = servicebroker.NewDeleteServiceBroker(ui, config, repoLocator.GetServiceBrokerRepository())
	factory.cmdsByName["delete-space"] = space.NewDeleteSpace(ui, config, repoLocator.GetSpaceRepository())
	factory.cmdsByName["delete-user"] = user.NewDeleteUser(ui, config, repoLocator.GetUserRepository())
	factory.cmdsByName["domains"] = domain.NewListDomains(ui, config, repoLocator.GetDomainRepository())
	factory.cmdsByName["env"] = application.NewEnv(ui, config, repoLocator.GetApplicationRepository())
	factory.cmdsByName["events"] = application.NewEvents(ui, config, repoLocator.GetAppEventsRepository())
	factory.cmdsByName["files"] = application.NewFiles(ui, config, repoLocator.GetAppFilesRepository())
	factory.cmdsByName["login"] = commands.NewLogin(ui, config, repoLocator.GetAuthenticationRepository(), repoLocator.GetEndpointRepository(), repoLocator.GetOrganizationRepository(), repoLocator.GetSpaceRepository())
	factory.cmdsByName["logout"] = commands.NewLogout(ui, config)
	factory.cmdsByName["logs"] = application.NewLogs(ui, config, repoLocator.GetLogsNoaaRepository())
	factory.cmdsByName["oauth-token"] = commands.NewOAuthToken(ui, config, repoLocator.GetAuthenticationRepository())
	factory.cmdsByName["org"] = organization.NewShowOrg(ui, config)
	factory.cmdsByName["org-users"] = user.NewOrgUsers(ui, config, repoLocator.GetUserRepository())
	factory.cmdsByName["orgs"] = organization.NewListOrgs(ui, config, repoLocator.GetOrganizationRepository())
	factory.cmdsByName["passwd"] = commands.NewPassword(ui, repoLocator.GetPasswordRepository(), config)
	factory.cmdsByName["purge-service-offering"] = service.NewPurgeServiceOffering(ui, config, repoLocator.GetServiceRepository())
	factory.cmdsByName["quotas"] = quota.NewListQuotas(ui, config, repoLocator.GetQuotaRepository())
	factory.cmdsByName["quota"] = quota.NewShowQuota(ui, config, repoLocator.GetQuotaRepository())
	factory.cmdsByName["create-quota"] = quota.NewCreateQuota(ui, config, repoLocator.GetQuotaRepository())
	factory.cmdsByName["update-quota"] = quota.NewUpdateQuota(ui, config, repoLocator.GetQuotaRepository())
	factory.cmdsByName["delete-quota"] = quota.NewDeleteQuota(ui, config, repoLocator.GetQuotaRepository())
	factory.cmdsByName["rename"] = application.NewRenameApp(ui, config, repoLocator.GetApplicationRepository())
	factory.cmdsByName["rename-buildpack"] = buildpack.NewRenameBuildpack(ui, repoLocator.GetBuildpackRepository())
	factory.cmdsByName["rename-org"] = organization.NewRenameOrg(ui, config, repoLocator.GetOrganizationRepository())
	factory.cmdsByName["rename-service"] = service.NewRenameService(ui, config, repoLocator.GetServiceRepository())
	factory.cmdsByName["rename-service-broker"] = servicebroker.NewRenameServiceBroker(ui, config, repoLocator.GetServiceBrokerRepository())
	factory.cmdsByName["rename-space"] = space.NewRenameSpace(ui, config, repoLocator.GetSpaceRepository())
	factory.cmdsByName["routes"] = route.NewListRoutes(ui, config, repoLocator.GetRouteRepository())
	factory.cmdsByName["check-route"] = route.NewCheckRoute(ui, config, repoLocator.GetRouteRepository(), repoLocator.GetDomainRepository())
	factory.cmdsByName["service"] = service.NewShowService(ui)
	factory.cmdsByName["service-auth-tokens"] = serviceauthtoken.NewListServiceAuthTokens(ui, config, repoLocator.GetServiceAuthTokenRepository())
	factory.cmdsByName["service-brokers"] = servicebroker.NewListServiceBrokers(ui, config, repoLocator.GetServiceBrokerRepository())
	factory.cmdsByName["services"] = service.NewListServices(ui, config, repoLocator.GetServiceSummaryRepository())
	factory.cmdsByName["migrate-service-instances"] = service.NewMigrateServiceInstances(ui, config, repoLocator.GetServiceRepository())
	factory.cmdsByName["set-env"] = application.NewSetEnv(ui, config, repoLocator.GetApplicationRepository())
	factory.cmdsByName["set-org-role"] = user.NewSetOrgRole(ui, config, repoLocator.GetUserRepository())
	factory.cmdsByName["set-quota"] = organization.NewSetQuota(ui, config, repoLocator.GetQuotaRepository())
	factory.cmdsByName["create-shared-domain"] = domain.NewCreateSharedDomain(ui, config, repoLocator.GetDomainRepository())
	factory.cmdsByName["space"] = space.NewShowSpace(ui, config, repoLocator.GetSpaceQuotaRepository())
	factory.cmdsByName["space-users"] = user.NewSpaceUsers(ui, config, repoLocator.GetSpaceRepository(), repoLocator.GetUserRepository())
	factory.cmdsByName["spaces"] = space.NewListSpaces(ui, config, repoLocator.GetSpaceRepository())
	factory.cmdsByName["stacks"] = commands.NewListStacks(ui, config, repoLocator.GetStackRepository())
	factory.cmdsByName["stack"] = commands.NewListStack(ui, config, repoLocator.GetStackRepository())
	factory.cmdsByName["target"] = commands.NewTarget(ui, config, repoLocator.GetOrganizationRepository(), repoLocator.GetSpaceRepository())
	factory.cmdsByName["unbind-service"] = service.NewUnbindService(ui, config, repoLocator.GetServiceBindingRepository())
	factory.cmdsByName["unset-env"] = application.NewUnsetEnv(ui, config, repoLocator.GetApplicationRepository())
	factory.cmdsByName["unset-org-role"] = user.NewUnsetOrgRole(ui, config, repoLocator.GetUserRepository())
	factory.cmdsByName["unset-space-role"] = user.NewUnsetSpaceRole(ui, config, repoLocator.GetSpaceRepository(), repoLocator.GetUserRepository())
	factory.cmdsByName["update-buildpack"] = buildpack.NewUpdateBuildpack(ui, repoLocator.GetBuildpackRepository(), repoLocator.GetBuildpackBitsRepository())
	factory.cmdsByName["update-service-broker"] = servicebroker.NewUpdateServiceBroker(ui, config, repoLocator.GetServiceBrokerRepository())
	factory.cmdsByName["update-service-auth-token"] = serviceauthtoken.NewUpdateServiceAuthToken(ui, config, repoLocator.GetServiceAuthTokenRepository())
	factory.cmdsByName["update-user-provided-service"] = service.NewUpdateUserProvidedService(ui, config, repoLocator.GetUserProvidedServiceInstanceRepository())
	factory.cmdsByName["create-security-group"] = securitygroup.NewCreateSecurityGroup(ui, config, repoLocator.GetSecurityGroupRepository())
	factory.cmdsByName["update-security-group"] = securitygroup.NewUpdateSecurityGroup(ui, config, repoLocator.GetSecurityGroupRepository())
	factory.cmdsByName["delete-security-group"] = securitygroup.NewDeleteSecurityGroup(ui, config, repoLocator.GetSecurityGroupRepository())
	factory.cmdsByName["security-group"] = securitygroup.NewShowSecurityGroup(ui, config, repoLocator.GetSecurityGroupRepository())
	factory.cmdsByName["security-groups"] = securitygroup.NewSecurityGroups(ui, config, repoLocator.GetSecurityGroupRepository())
	factory.cmdsByName["bind-staging-security-group"] = securitygroup.NewBindToStagingGroup(
		ui,
		config,
		repoLocator.GetSecurityGroupRepository(),
		repoLocator.GetStagingSecurityGroupsRepository(),
	)
	factory.cmdsByName["staging-security-groups"] = securitygroup.NewListStagingSecurityGroups(ui, config, repoLocator.GetStagingSecurityGroupsRepository())
	factory.cmdsByName["unbind-staging-security-group"] = securitygroup.NewUnbindFromStagingGroup(
		ui,
		config,
		repoLocator.GetSecurityGroupRepository(),
		repoLocator.GetStagingSecurityGroupsRepository(),
	)
	factory.cmdsByName["bind-running-security-group"] = securitygroup.NewBindToRunningGroup(
		ui,
		config,
		repoLocator.GetSecurityGroupRepository(),
		repoLocator.GetRunningSecurityGroupsRepository(),
	)
	factory.cmdsByName["unbind-running-security-group"] = securitygroup.NewUnbindFromRunningGroup(
		ui,
		config,
		repoLocator.GetSecurityGroupRepository(),
		repoLocator.GetRunningSecurityGroupsRepository(),
	)
	factory.cmdsByName["running-security-groups"] = securitygroup.NewListRunningSecurityGroups(ui, config, repoLocator.GetRunningSecurityGroupsRepository())
	factory.cmdsByName["bind-security-group"] = securitygroup.NewBindSecurityGroup(
		ui,
		config,
		repoLocator.GetSecurityGroupRepository(),
		repoLocator.GetSpaceRepository(),
		repoLocator.GetOrganizationRepository(),
		repoLocator.GetSecurityGroupSpaceBinder(),
	)
	factory.cmdsByName["unbind-security-group"] = securitygroup.NewUnbindSecurityGroup(ui, config, repoLocator.GetSecurityGroupRepository(), repoLocator.GetOrganizationRepository(), repoLocator.GetSpaceRepository(), repoLocator.GetSecurityGroupSpaceBinder())

	createRoute := route.NewCreateRoute(ui, config, repoLocator.GetRouteRepository())
	factory.cmdsByName["create-route"] = createRoute
	factory.cmdsByName["map-route"] = route.NewMapRoute(ui, config, repoLocator.GetRouteRepository(), createRoute)
	factory.cmdsByName["unmap-route"] = route.NewUnmapRoute(ui, config, repoLocator.GetRouteRepository())

	displayApp := application.NewShowApp(ui, config, repoLocator.GetAppSummaryRepository(), repoLocator.GetAppInstancesRepository(), repoLocator.GetLogsNoaaRepository())
	start := application.NewStart(ui, config, displayApp, repoLocator.GetApplicationRepository(), repoLocator.GetAppInstancesRepository(), repoLocator.GetLogsNoaaRepository())
	stop := application.NewStop(ui, config, repoLocator.GetApplicationRepository())
	restart := application.NewRestart(ui, config, start, stop)
	restage := application.NewRestage(ui, config, repoLocator.GetApplicationRepository(), start)
	bind := service.NewBindService(ui, config, repoLocator.GetServiceBindingRepository())

	factory.cmdsByName["app"] = displayApp
	factory.cmdsByName["bind-service"] = bind
	factory.cmdsByName["start"] = start
	factory.cmdsByName["stop"] = stop
	factory.cmdsByName["restart"] = restart
	factory.cmdsByName["restart-app-instance"] = application.NewRestartAppInstance(ui, config, repoLocator.GetAppInstancesRepository())
	factory.cmdsByName["restage"] = restage
	factory.cmdsByName["push"] = application.NewPush(
		ui, config, manifestRepo, start, stop, bind,
		repoLocator.GetApplicationRepository(),
		repoLocator.GetDomainRepository(),
		repoLocator.GetRouteRepository(),
		repoLocator.GetStackRepository(),
		repoLocator.GetServiceRepository(),
		repoLocator.GetAuthenticationRepository(),
		generator.NewWordGenerator(),
		actors.NewPushActor(repoLocator.GetApplicationBitsRepository(), app_files.ApplicationZipper{}, app_files.ApplicationFiles{}),
		app_files.ApplicationZipper{},
		app_files.ApplicationFiles{})

	factory.cmdsByName["scale"] = application.NewScale(ui, config, restart, repoLocator.GetApplicationRepository())

	spaceRoleSetter := user.NewSetSpaceRole(ui, config, repoLocator.GetSpaceRepository(), repoLocator.GetUserRepository())
	factory.cmdsByName["set-space-role"] = spaceRoleSetter
	factory.cmdsByName["create-space"] = space.NewCreateSpace(ui, config, spaceRoleSetter, repoLocator.GetSpaceRepository(), repoLocator.GetOrganizationRepository(), repoLocator.GetUserRepository(), repoLocator.GetSpaceQuotaRepository())

	factory.cmdsByName["service-access"] = serviceaccess.NewServiceAccess(
		ui, config,
		actors.NewServiceHandler(
			repoLocator.GetOrganizationRepository(),
			brokerBuilder,
			serviceBuilder,
		),
		repoLocator.GetAuthenticationRepository(),
	)
	factory.cmdsByName["enable-service-access"] = serviceaccess.NewEnableServiceAccess(
		ui, config,
		actors.NewServicePlanHandler(
			repoLocator.GetServicePlanRepository(),
			repoLocator.GetServicePlanVisibilityRepository(),
			repoLocator.GetOrganizationRepository(),
			planBuilder,
			serviceBuilder,
		),
		repoLocator.GetAuthenticationRepository(),
	)
	factory.cmdsByName["disable-service-access"] = serviceaccess.NewDisableServiceAccess(
		ui, config,
		actors.NewServicePlanHandler(
			repoLocator.GetServicePlanRepository(),
			repoLocator.GetServicePlanVisibilityRepository(),
			repoLocator.GetOrganizationRepository(),
			planBuilder,
			serviceBuilder,
		),
		repoLocator.GetAuthenticationRepository(),
	)

	factory.cmdsByName["marketplace"] = service.NewMarketplaceServices(ui, config, serviceBuilder)

	factory.cmdsByName["create-space-quota"] = spacequota.NewCreateSpaceQuota(ui, config, repoLocator.GetSpaceQuotaRepository(), repoLocator.GetOrganizationRepository())
	factory.cmdsByName["delete-space-quota"] = spacequota.NewDeleteSpaceQuota(ui, config, repoLocator.GetSpaceQuotaRepository())

	factory.cmdsByName["space-quotas"] = spacequota.NewListSpaceQuotas(ui, config, repoLocator.GetSpaceQuotaRepository())
	factory.cmdsByName["space-quota"] = spacequota.NewSpaceQuota(ui, config, repoLocator.GetSpaceQuotaRepository())
	factory.cmdsByName["update-space-quota"] = spacequota.NewUpdateSpaceQuota(ui, config, repoLocator.GetSpaceQuotaRepository())
	factory.cmdsByName["set-space-quota"] = spacequota.NewSetSpaceQuota(ui, config, repoLocator.GetSpaceRepository(), repoLocator.GetSpaceQuotaRepository())
	factory.cmdsByName["unset-space-quota"] = spacequota.NewUnsetSpaceQuota(ui, config, repoLocator.GetSpaceQuotaRepository(), repoLocator.GetSpaceRepository())
	factory.cmdsByName["feature-flags"] = featureflag.NewListFeatureFlags(ui, config, repoLocator.GetFeatureFlagRepository())
	factory.cmdsByName["feature-flag"] = featureflag.NewShowFeatureFlag(ui, config, repoLocator.GetFeatureFlagRepository())
	factory.cmdsByName["enable-feature-flag"] = featureflag.NewEnableFeatureFlag(ui, config, repoLocator.GetFeatureFlagRepository())
	factory.cmdsByName["disable-feature-flag"] = featureflag.NewDisableFeatureFlag(ui, config, repoLocator.GetFeatureFlagRepository())
	factory.cmdsByName["running-environment-variable-group"] = environmentvariablegroup.NewRunningEnvironmentVariableGroup(ui, config, repoLocator.GetEnvironmentVariableGroupsRepository())
	factory.cmdsByName["staging-environment-variable-group"] = environmentvariablegroup.NewStagingEnvironmentVariableGroup(ui, config, repoLocator.GetEnvironmentVariableGroupsRepository())
	factory.cmdsByName["set-staging-environment-variable-group"] = environmentvariablegroup.NewSetStagingEnvironmentVariableGroup(ui, config, repoLocator.GetEnvironmentVariableGroupsRepository())
	factory.cmdsByName["set-running-environment-variable-group"] = environmentvariablegroup.NewSetRunningEnvironmentVariableGroup(ui, config, repoLocator.GetEnvironmentVariableGroupsRepository())

	factory.cmdsByName["uninstall-plugin"] = plugin.NewPluginUninstall(ui, pluginConfig, rpcService)
	factory.cmdsByName["install-plugin"] = plugin.NewPluginInstall(ui, config, pluginConfig, factory.cmdsByName, actor_plugin_repo.NewPluginRepo(), utils.NewSha1Checksum(""), rpcService)
	factory.cmdsByName["plugins"] = plugin.NewPlugins(ui, pluginConfig)

	factory.cmdsByName["add-plugin-repo"] = plugin_repo.NewAddPluginRepo(ui, config)
	factory.cmdsByName["list-plugin-repos"] = plugin_repo.NewListPluginRepos(ui, config)
	factory.cmdsByName["remove-plugin-repo"] = plugin_repo.NewRemovePluginRepo(ui, config)
	factory.cmdsByName["repo-plugins"] = plugin_repo.NewRepoPlugins(ui, config, actor_plugin_repo.NewPluginRepo())

	factory.cmdsByName["copy-source"] = application.NewCopySource(
		ui,
		config,
		repoLocator.GetAuthenticationRepository(),
		repoLocator.GetApplicationRepository(),
		repoLocator.GetOrganizationRepository(),
		repoLocator.GetSpaceRepository(),
		repoLocator.GetCopyApplicationSourceRepository(),
		restart, //note this is built up above.
	)

	factory.cmdsByName["share-private-domain"] = organization.NewSharePrivateDomain(ui, config, repoLocator.GetOrganizationRepository(), repoLocator.GetDomainRepository())
	factory.cmdsByName["unshare-private-domain"] = organization.NewUnsharePrivateDomain(ui, config, repoLocator.GetOrganizationRepository(), repoLocator.GetDomainRepository())

	return
}

func (f concreteFactory) GetByCmdName(cmdName string) (cmd command.Command, err error) {
	cmd, found := f.cmdsByName[cmdName]
	if !found {
		for _, c := range f.cmdsByName {
			if c.Metadata().ShortName == cmdName {
				return c, nil
			}
		}

		err = errors.New(T("Command not found"))
	}
	return
}

func (f concreteFactory) CheckIfCoreCmdExists(cmdName string) bool {
	if _, exists := f.cmdsByName[cmdName]; exists {
		return true
	}

	for _, singleCmd := range f.cmdsByName {
		metaData := singleCmd.Metadata()
		if metaData.ShortName == cmdName {
			return true
		}
	}

	return false
}

func (factory concreteFactory) CommandMetadatas() (commands []command_metadata.CommandMetadata) {
	for _, command := range factory.cmdsByName {
		commands = append(commands, command.Metadata())
	}
	return
}

func (f concreteFactory) GetCommandFlags(command string) []string {
	cmd, err := f.GetByCmdName(command)
	if err != nil {
		return []string{}
	}

	var flags []string
	for _, flag := range cmd.Metadata().Flags {
		switch t := flag.(type) {
		default:
		case flag_helpers.StringSliceFlagWithNoDefault:
			flags = append(flags, t.Name)
		case flag_helpers.IntFlagWithNoDefault:
			flags = append(flags, t.Name)
		case flag_helpers.StringFlagWithNoDefault:
			flags = append(flags, t.Name)
		case cli.BoolFlag:
			flags = append(flags, t.Name)
		}
	}

	return flags
}

func (f concreteFactory) GetCommandTotalArgs(command string) (int, error) {
	cmd, err := f.GetByCmdName(command)
	if err != nil {
		return 0, err
	}

	return cmd.Metadata().TotalArgs, nil
}
