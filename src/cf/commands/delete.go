package commands

import (
	"cf/api"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
	"strings"
)

type Delete struct {
	ui      terminal.UI
	appRepo api.ApplicationRepository
	appReq  requirements.ApplicationRequirement
}

func NewDelete(ui terminal.UI, appRepo api.ApplicationRepository) (cmd *Delete) {
	cmd = new(Delete)
	cmd.ui = ui
	cmd.appRepo = appRepo
	return
}

func (cmd *Delete) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) == 0 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "delete")
		return
	}

	return
}

func (cmd *Delete) Run(c *cli.Context) {
	appName := c.Args()[0]
	force := c.Bool("f")

	if !force {
		response := strings.ToLower(cmd.ui.Ask(
			"Really delete %s?%s",
			terminal.EntityNameColor(appName),
			terminal.PromptColor(">"),
		))
		if response != "y" && response != "yes" {
			return
		}
	}

	cmd.ui.Say("Deleting app %s...", terminal.EntityNameColor(appName))

	app, found, apiErr := cmd.appRepo.FindByName(appName)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Message)
		return
	}

	if !found {
		cmd.ui.Ok()
		cmd.ui.Warn("App %s does not exist.", appName)
		return
	}

	apiErr = cmd.appRepo.Delete(app)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Ok()
	return
}
