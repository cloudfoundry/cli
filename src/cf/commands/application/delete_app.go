package application

import (
	"cf/api"
	"cf/configuration"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type DeleteApp struct {
	ui      terminal.UI
	config  *configuration.Configuration
	appRepo api.ApplicationRepository
	appReq  requirements.ApplicationRequirement
}

func NewDeleteApp(ui terminal.UI, config *configuration.Configuration, appRepo api.ApplicationRepository) (cmd *DeleteApp) {
	cmd = new(DeleteApp)
	cmd.ui = ui
	cmd.config = config
	cmd.appRepo = appRepo
	return
}

func (cmd *DeleteApp) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) == 0 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "delete")
		return
	}

	return
}

func (cmd *DeleteApp) Run(c *cli.Context) {
	appName := c.Args()[0]
	force := c.Bool("f")

	if !force {
		response := cmd.ui.Confirm(
			"Really delete %s?%s",
			terminal.EntityNameColor(appName),
			terminal.PromptColor(">"),
		)
		if !response {
			return
		}
	}

	cmd.ui.Say("Deleting app %s in org %s / space %s as %s...",
		terminal.EntityNameColor(appName),
		terminal.EntityNameColor(cmd.config.OrganizationFields.Name),
		terminal.EntityNameColor(cmd.config.SpaceFields.Name),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	app, apiResponse := cmd.appRepo.FindByName(appName)

	if apiResponse.IsError() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	if apiResponse.IsNotFound() {
		cmd.ui.Ok()
		cmd.ui.Warn("App %s does not exist.", appName)
		return
	}

	apiResponse = cmd.appRepo.Delete(app.Guid)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Ok()
	return
}
