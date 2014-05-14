package application

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type Env struct {
	ui      terminal.UI
	config  configuration.Reader
	appRepo api.ApplicationRepository
}

func NewEnv(ui terminal.UI, config configuration.Reader, appRepo api.ApplicationRepository) (cmd *Env) {
	cmd = new(Env)
	cmd.ui = ui
	cmd.config = config
	cmd.appRepo = appRepo
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

func (cmd *Env) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) ([]requirements.Requirement, error) {
	if len(c.Args()) != 1 {
		cmd.ui.FailWithUsage(c)
	}

	return []requirements.Requirement{requirementsFactory.NewLoginRequirement()}, nil
}

func (cmd *Env) Run(c *cli.Context) {
	app, err := cmd.appRepo.Read(c.Args()[0])
	if notFound, ok := err.(*errors.ModelNotFoundError); ok {
		cmd.ui.Failed(notFound.Error())
	}

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
