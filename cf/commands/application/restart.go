package application

import (
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
)

//go:generate counterfeiter -o fakes/fake_application_restarter.go . ApplicationRestarter
type ApplicationRestarter interface {
	command_registry.Command
	ApplicationRestart(app models.Application, orgName string, spaceName string)
}

type Restart struct {
	ui      terminal.UI
	config  core_config.Reader
	starter ApplicationStarter
	stopper ApplicationStopper
	appReq  requirements.ApplicationRequirement
}

func init() {
	command_registry.Register(&Restart{})
}

func (cmd *Restart) MetaData() command_registry.CommandMetadata {
	return command_registry.CommandMetadata{
		Name:        "restart",
		ShortName:   "rs",
		Description: T("Restart an app"),
		Usage:       T("CF_NAME restart APP_NAME"),
	}
}

func (cmd *Restart) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) (reqs []requirements.Requirement, err error) {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + command_registry.Commands.CommandUsage("restart"))
	}

	cmd.appReq = requirementsFactory.NewApplicationRequirement(fc.Args()[0])

	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
		cmd.appReq,
	}
	return
}

func (cmd *Restart) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config

	//get start for dependency
	starter := command_registry.Commands.FindCommand("start")
	starter = starter.SetDependency(deps, false)
	cmd.starter = starter.(ApplicationStarter)

	//get stop for dependency
	stopper := command_registry.Commands.FindCommand("stop")
	stopper = stopper.SetDependency(deps, false)
	cmd.stopper = stopper.(ApplicationStopper)

	return cmd
}

func (cmd *Restart) Execute(c flags.FlagContext) {
	app := cmd.appReq.GetApplication()
	cmd.ApplicationRestart(app, cmd.config.OrganizationFields().Name, cmd.config.SpaceFields().Name)
}

func (cmd *Restart) ApplicationRestart(app models.Application, orgName, spaceName string) {
	stoppedApp, err := cmd.stopper.ApplicationStop(app, orgName, spaceName)
	if err != nil {
		cmd.ui.Failed(err.Error())
		return
	}

	cmd.ui.Say("")

	_, err = cmd.starter.ApplicationStart(stoppedApp, orgName, spaceName)
	if err != nil {
		cmd.ui.Failed(err.Error())
		return
	}
}
