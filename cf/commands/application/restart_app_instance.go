package application

import (
	"strconv"

	"github.com/cloudfoundry/cli/cf/api/app_instances"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type RestartAppInstance struct {
	ui               terminal.UI
	config           core_config.Reader
	appReq           requirements.ApplicationRequirement
	appInstancesRepo app_instances.AppInstancesRepository
}

func NewRestartAppInstance(ui terminal.UI, config core_config.Reader, appInstancesRepo app_instances.AppInstancesRepository) (cmd *RestartAppInstance) {
	cmd = new(RestartAppInstance)
	cmd.ui = ui
	cmd.config = config
	cmd.appInstancesRepo = appInstancesRepo
	return
}

func (cmd *RestartAppInstance) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "restart-app-instance",
		Description: T("Terminate the running application Instance at the given index and instantiate a new instance of the application with the same index"),
		Usage:       T("CF_NAME restart-app-instance APP_NAME INDEX"),
	}
}

func (cmd *RestartAppInstance) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 2 {
		cmd.ui.FailWithUsage(c)
	}

	if cmd.appReq == nil {
		cmd.appReq = requirementsFactory.NewApplicationRequirement(c.Args()[0])
	} else {
		cmd.appReq.SetApplicationName(c.Args()[0])
	}

	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
		cmd.appReq,
	}

	return
}

func (cmd *RestartAppInstance) Run(c *cli.Context) {
	app := cmd.appReq.GetApplication()

	instance, err := strconv.Atoi(c.Args()[1])

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
