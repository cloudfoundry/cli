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

func NewDelete(ui terminal.UI, appRepo api.ApplicationRepository) (d *Delete) {
	d = new(Delete)
	d.ui = ui
	d.appRepo = appRepo
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

func (d *Delete) Run(c *cli.Context) {
	appName := c.Args()[0]
	force := c.Bool("f")

	if !force {
		response := strings.ToLower(d.ui.Ask(
			"Really delete %s?%s",
			terminal.EntityNameColor(appName),
			terminal.PromptColor(">"),
		))
		if response != "y" && response != "yes" {
			return
		}
	}

	d.ui.Say("Deleting app %s...", terminal.EntityNameColor(appName))

	app, found, apiErr := d.appRepo.FindByName(appName)
	if apiErr != nil {
		d.ui.Failed(apiErr.Message)
		return
	}

	if !found {
		d.ui.Ok()
		d.ui.Warn("App %s does not exist.", appName)
		return
	}

	apiErr = d.appRepo.Delete(app)
	if apiErr != nil {
		d.ui.Failed(apiErr.Error())
		return
	}

	d.ui.Ok()
	return
}
