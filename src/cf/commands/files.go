package commands

import (
	"cf/api"
	"cf/requirements"
	"cf/terminal"
	"github.com/codegangsta/cli"
)

type Files struct {
	ui           terminal.UI
	appFilesRepo api.AppFilesRepository
	appReq       requirements.ApplicationRequirement
}

func NewFiles(ui terminal.UI, appFilesRepo api.AppFilesRepository) (cmd *Files) {
	cmd = new(Files)
	cmd.ui = ui
	cmd.appFilesRepo = appFilesRepo
	return
}

func (cmd *Files) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	cmd.appReq = reqFactory.NewApplicationRequirement(c.String("app"))

	reqs = []requirements.Requirement{
		reqFactory.NewLoginRequirement(),
		reqFactory.NewSpaceRequirement(),
		cmd.appReq,
	}
	return
}

func (cmd *Files) Run(c *cli.Context) {
	cmd.ui.Say("Getting files...")

	app := cmd.appReq.GetApplication()
	path := c.String("path")

	list, err := cmd.appFilesRepo.ListFiles(app, path)
	if err != nil {
		cmd.ui.Failed("", err)
		return
	}

	cmd.ui.Ok()
	cmd.ui.Say(list)
}
