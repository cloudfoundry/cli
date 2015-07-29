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
	"github.com/cloudfoundry/cli/cf/commands/application"
	"github.com/cloudfoundry/cli/cf/commands/plugin"
	"github.com/cloudfoundry/cli/cf/commands/service"
	"github.com/cloudfoundry/cli/cf/commands/serviceaccess"
	"github.com/cloudfoundry/cli/cf/commands/servicekey"
	"github.com/cloudfoundry/cli/cf/commands/space"
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

	factory.cmdsByName["create-user"] = user.NewCreateUser(ui, config, repoLocator.GetUserRepository())
	factory.cmdsByName["delete-service-key"] = servicekey.NewDeleteServiceKey(ui, config, repoLocator.GetServiceRepository(), repoLocator.GetServiceKeyRepository())
	factory.cmdsByName["service-key"] = servicekey.NewGetServiceKey(ui, config, repoLocator.GetServiceRepository(), repoLocator.GetServiceKeyRepository())
	factory.cmdsByName["unbind-service"] = service.NewUnbindService(ui, config, repoLocator.GetServiceBindingRepository())
	factory.cmdsByName["update-user-provided-service"] = service.NewUpdateUserProvidedService(ui, config, repoLocator.GetUserProvidedServiceInstanceRepository())

	displayApp := application.NewShowApp(ui, config, repoLocator.GetAppSummaryRepository(), repoLocator.GetAppInstancesRepository(), repoLocator.GetLogsNoaaRepository())
	start := application.NewStart(ui, config, displayApp, repoLocator.GetApplicationRepository(), repoLocator.GetAppInstancesRepository(), repoLocator.GetLogsNoaaRepository(), repoLocator.GetOldLogsRepository())
	stop := application.NewStop(ui, config, repoLocator.GetApplicationRepository())
	restart := application.NewRestart(ui, config, start, stop)
	restage := application.NewRestage(ui, config, repoLocator.GetApplicationRepository(), start)
	bind := service.NewBindService(ui, config, repoLocator.GetServiceBindingRepository())

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

	factory.cmdsByName["install-plugin"] = plugin.NewPluginInstall(ui, config, pluginConfig, factory.cmdsByName, actor_plugin_repo.NewPluginRepo(), utils.NewSha1Checksum(""), rpcService)

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
