package application

import (
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
)

//go:generate counterfeiter . ApplicationRestarter

type ApplicationRestarter interface {
	commandregistry.Command
	ApplicationRestart(app models.Application, orgName string, spaceName string)
}

type Restart struct {
	ui      terminal.UI
	config  coreconfig.Reader
	starter ApplicationStarter
	stopper ApplicationStopper
	appReq  requirements.ApplicationRequirement
}

func init() {
	commandregistry.Register(&Restart{})
}

func (cmd *Restart) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "restart",
		ShortName:   "rs",
		Description: T("Restart an app"),
		Usage: []string{
			T("CF_NAME restart APP_NAME"),
		},
	}
}

func (cmd *Restart) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + commandregistry.Commands.CommandUsage("restart"))
	}

	cmd.appReq = requirementsFactory.NewApplicationRequirement(fc.Args()[0])

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
		cmd.appReq,
	}

	return reqs
}

func (cmd *Restart) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config

	//get start for dependency
	starter := commandregistry.Commands.FindCommand("start")
	starter = starter.SetDependency(deps, false)
	cmd.starter = starter.(ApplicationStarter)

	//get stop for dependency
	stopper := commandregistry.Commands.FindCommand("stop")
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
