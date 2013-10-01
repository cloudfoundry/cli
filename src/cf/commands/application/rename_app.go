package application

import (
	"cf/api"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type RenameApp struct {
	ui      terminal.UI
	appRepo api.ApplicationRepository
	appReq  requirements.ApplicationRequirement
}

func NewRenameApp(ui terminal.UI, appRepo api.ApplicationRepository) (cmd *RenameApp) {
	cmd = new(RenameApp)
	cmd.ui = ui
	cmd.appRepo = appRepo
	return
}

func (cmd *RenameApp) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 2 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "rename")
		return
	}
	cmd.appReq = reqFactory.NewApplicationRequirement(c.Args()[0])
	reqs = []requirements.Requirement{
		reqFactory.NewLoginRequirement(),
		cmd.appReq,
	}
	return
}

func (cmd *RenameApp) Run(c *cli.Context) {
	app := cmd.appReq.GetApplication()
	new_name := c.Args()[1]
	cmd.ui.Say("Renaming %s to %s...", terminal.EntityNameColor(app.Name), terminal.EntityNameColor(new_name))

	err := cmd.appRepo.Rename(app, new_name)
	if err != nil {
		cmd.ui.Failed(err.Error())
		return
	}
	cmd.ui.Ok()
}
