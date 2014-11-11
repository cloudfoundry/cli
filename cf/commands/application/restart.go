package application

import (
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type Restart struct {
	ui      terminal.UI
	config  core_config.Reader
	starter ApplicationStarter
	stopper ApplicationStopper
	appReq  requirements.ApplicationRequirement
}

type ApplicationRestarter interface {
	ApplicationRestart(app models.Application, orgName string, spaceName string)
}

func NewRestart(ui terminal.UI, config core_config.Reader, starter ApplicationStarter, stopper ApplicationStopper) (cmd *Restart) {
	cmd = new(Restart)
	cmd.ui = ui
	cmd.config = config
	cmd.starter = starter
	cmd.stopper = stopper
	return
}

func (cmd *Restart) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "restart",
		ShortName:   "rs",
		Description: T("Restart an app"),
		Usage:       T("CF_NAME restart APP"),
	}
}

func (cmd *Restart) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) == 0 {
		cmd.ui.FailWithUsage(c)
	}

	cmd.appReq = requirementsFactory.NewApplicationRequirement(c.Args()[0])

	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
		cmd.appReq,
	}
	return
}

func (cmd *Restart) Run(c *cli.Context) {
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
