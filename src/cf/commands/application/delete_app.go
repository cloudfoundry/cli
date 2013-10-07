package application

import (
	"cf/api"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type DeleteApp struct {
	ui      terminal.UI
	appRepo api.ApplicationRepository
	appReq  requirements.ApplicationRequirement
}

func NewDeleteApp(ui terminal.UI, appRepo api.ApplicationRepository) (cmd *DeleteApp) {
	cmd = new(DeleteApp)
	cmd.ui = ui
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

	cmd.ui.Say("Deleting app %s...", terminal.EntityNameColor(appName))

	app, apiStatus := cmd.appRepo.FindByName(appName)

	if apiStatus.IsError() {
		cmd.ui.Failed(apiStatus.Message)
		return
	}

	if apiStatus.IsNotFound() {
		cmd.ui.Ok()
		cmd.ui.Warn("App %s does not exist.", appName)
		return
	}

	apiStatus = cmd.appRepo.Delete(app)
	if apiStatus.IsNotSuccessful() {
		cmd.ui.Failed(apiStatus.Message)
		return
	}

	cmd.ui.Ok()
	return
}
