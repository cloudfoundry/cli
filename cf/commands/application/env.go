package application

import (
	"errors"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type Env struct {
	ui     terminal.UI
	config configuration.Reader
	appReq requirements.ApplicationRequirement
}

func NewEnv(ui terminal.UI, config configuration.Reader) (cmd *Env) {
	cmd = new(Env)
	cmd.ui = ui
	cmd.config = config
	return
}

func (command *Env) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "env",
		ShortName:   "e",
		Description: "Show all env variables for an app",
		Usage:       "CF_NAME env APP",
	}
}

func (cmd *Env) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) < 1 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "env")
		return
	}

	cmd.appReq = requirementsFactory.NewApplicationRequirement(c.Args()[0])

	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		cmd.appReq,
	}
	return
}

func (cmd *Env) Run(c *cli.Context) {
	app := cmd.appReq.GetApplication()

	cmd.ui.Say("Getting env variables for app %s in org %s / space %s as %s...",
		terminal.EntityNameColor(app.Name),
		terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
		terminal.EntityNameColor(cmd.config.SpaceFields().Name),
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
