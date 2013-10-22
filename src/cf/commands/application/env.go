package application

import (
	"cf/configuration"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type Env struct {
	ui     terminal.UI
	config *configuration.Configuration
	appReq requirements.ApplicationRequirement
}

func NewEnv(ui terminal.UI, config *configuration.Configuration) (cmd *Env) {
	cmd = new(Env)
	cmd.ui = ui
	cmd.config = config
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

	cmd.ui.Say("Getting env variables for app %s in org %s / space %s as %s...",
		terminal.EntityNameColor(app.Name),
		terminal.EntityNameColor(cmd.config.Organization.Name),
		terminal.EntityNameColor(cmd.config.Space.Name),
		terminal.EntityNameColor(cmd.config.Username()),
	)
	envVars := app.EnvironmentVars

	cmd.ui.Ok()
	cmd.ui.Say("")

	if len(envVars) == 0 {
		cmd.ui.Say("No env variables exist")
		return
	}
	for key, value := range envVars {
		cmd.ui.Say("%s: %s", key, terminal.EntityNameColor(value))
	}
}
