package application

import (
	"strconv"

	"github.com/cloudfoundry/cli/cf/api/app_instances"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
)

type RestartAppInstance struct {
	ui               terminal.UI
	config           core_config.Reader
	appReq           requirements.ApplicationRequirement
	appInstancesRepo app_instances.AppInstancesRepository
}

func init() {
	command_registry.Register(&RestartAppInstance{})
}

func (cmd *RestartAppInstance) MetaData() command_registry.CommandMetadata {
	return command_registry.CommandMetadata{
		Name:        "restart-app-instance",
		Description: T("Terminate the running application Instance at the given index and instantiate a new instance of the application with the same index"),
		Usage:       T("CF_NAME restart-app-instance APP_NAME INDEX"),
	}
}

func (cmd *RestartAppInstance) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) (reqs []requirements.Requirement, err error) {
	if len(fc.Args()) != 2 {
		usage := command_registry.Commands.CommandUsage("restart-app-instance")
		cmd.ui.Failed(T("Incorrect Usage. Requires arguments\n\n") + usage)
	}

	appName := fc.Args()[0]

	cmd.appReq = requirementsFactory.NewApplicationRequirement(appName)

	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
		cmd.appReq,
	}

	return
}

func (cmd *RestartAppInstance) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.appInstancesRepo = deps.RepoLocator.GetAppInstancesRepository()
	return cmd
}

func (cmd *RestartAppInstance) Execute(fc flags.FlagContext) {
	app := cmd.appReq.GetApplication()

	instance, err := strconv.Atoi(fc.Args()[1])

	if err != nil {
		cmd.ui.Failed(T("Instance must be a non-negative integer"))
	}

	cmd.ui.Say(T("Restarting instance {{.Instance}} of application {{.AppName}} as {{.Username}}",
		map[string]interface{}{
			"Instance": instance,
			"AppName":  terminal.EntityNameColor(app.Name),
			"Username": terminal.EntityNameColor(cmd.config.Username()),
		}))

	err = cmd.appInstancesRepo.DeleteInstance(app.Guid, instance)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	cmd.ui.Ok()
	cmd.ui.Say("")
}
