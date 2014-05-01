package application

import (
	"errors"
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type ApplicationStopper interface {
	ApplicationStop(app models.Application) (updatedApp models.Application, err error)
}

type Stop struct {
	ui      terminal.UI
	config  configuration.Reader
	appRepo api.ApplicationRepository
	appReq  requirements.ApplicationRequirement
}

func NewStop(ui terminal.UI, config configuration.Reader, appRepo api.ApplicationRepository) (cmd *Stop) {
	cmd = new(Stop)
	cmd.ui = ui
	cmd.config = config
	cmd.appRepo = appRepo

	return
}

func (command *Stop) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "stop",
		ShortName:   "sp",
		Description: "Stop an app",
		Usage:       "CF_NAME stop APP",
	}
}

func (cmd *Stop) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) == 0 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "stop")
		return
	}

	cmd.appReq = requirementsFactory.NewApplicationRequirement(c.Args()[0])

	reqs = []requirements.Requirement{requirementsFactory.NewLoginRequirement(), cmd.appReq}
	return
}

func (cmd *Stop) ApplicationStop(app models.Application) (updatedApp models.Application, err error) {
	if app.State == "stopped" {
		updatedApp = app
		return
	}

	cmd.ui.Say("Stopping app %s in org %s / space %s as %s...",
		terminal.EntityNameColor(app.Name),
		terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
		terminal.EntityNameColor(cmd.config.SpaceFields().Name),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	state := "STOPPED"
	updatedApp, apiErr := cmd.appRepo.Update(app.Guid, models.AppParams{State: &state})
	if apiErr != nil {
		err = errors.New(apiErr.Error())
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Ok()
	return
}

func (cmd *Stop) Run(c *cli.Context) {
	app := cmd.appReq.GetApplication()
	if app.State == "stopped" {
		cmd.ui.Say(terminal.WarningColor("App " + app.Name + " is already stopped"))
	} else {
		cmd.ApplicationStop(app)
	}
}
