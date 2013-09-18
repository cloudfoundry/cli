package commands

import (
	"cf/api"
	"cf/requirements"
	term "cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type Rename struct {
	ui      term.UI
	appRepo api.ApplicationRepository
	appReq  requirements.ApplicationRequirement
}

func NewRename(ui term.UI, appRepo api.ApplicationRepository) (cmd *Rename) {
	cmd = new(Rename)
	cmd.ui = ui
	cmd.appRepo = appRepo
	return
}

func (cmd *Rename) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
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

func (cmd *Rename) Run(c *cli.Context) {
	app := cmd.appReq.GetApplication()
	new_name := c.Args()[1]
	cmd.ui.Say("Renaming %s to %s...", term.EntityNameColor(app.Name), term.EntityNameColor(new_name))

	err := cmd.appRepo.Rename(app, new_name)
	if err != nil {
		cmd.ui.Failed(err.Error())
		return
	}
	cmd.ui.Ok()
}
