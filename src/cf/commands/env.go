package commands

import (
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type Env struct {
	ui     terminal.UI
	appReq requirements.ApplicationRequirement
}

func NewEnv(ui terminal.UI) (cmd *Env) {
	cmd = new(Env)
	cmd.ui = ui
	return
}

func (cmd *Env) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) < 1 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "env")
		return
	}

	cmd.appReq = reqFactory.NewApplicationRequirement(c.Args()[0])

	reqs = []requirements.Requirement{
		reqFactory.NewLoginRequirement(),
		cmd.appReq,
	}
	return
}

func (cmd *Env) Run(c *cli.Context) {
	app := cmd.appReq.GetApplication()

	cmd.ui.Say("Getting env variables for %s...", terminal.EntityNameColor(app.Name))
	envVars := app.EnvironmentVars

	cmd.ui.Ok()
	if len(envVars) == 0 {
		cmd.ui.Say("No env variables exist")
		return
	}
	for key, value := range envVars {
		cmd.ui.Say("%s: %s", key, terminal.EntityNameColor(value))
	}
}
