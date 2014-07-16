package application

import (
	. "github.com/cloudfoundry/cli/cf/i18n"
	"strings"

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

func (cmd *Env) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "env",
		ShortName:   "e",
		Description: T("Show all env variables for an app"),
		Usage:       T("CF_NAME env APP"),
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

	cmd.ui.Say(T("Getting env variables for app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...",
		map[string]interface{}{
			"AppName":   terminal.EntityNameColor(app.Name),
			"OrgName":   terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
			"SpaceName": terminal.EntityNameColor(cmd.config.SpaceFields().Name),
			"Username":  terminal.EntityNameColor(cmd.config.Username())}))

	envVars, vcapServices, err := cmd.appRepo.ReadEnv(app.Guid)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	cmd.ui.Ok()
	cmd.ui.Say("")

	if len(vcapServices) > 0 {
		cmd.ui.Say(T("System-Provided:"))
		for _, line := range strings.Split(vcapServices, "\n") {
			cmd.ui.Say(line)
		}
	} else {
		cmd.ui.Say(T("No system-provided env variables have been set"))
	}

	if len(envVars) == 0 {
		cmd.ui.Say(T("No user-defined env variables have been set"))
		return
	}

	cmd.ui.Say(T("User-Provided:"))
	for key, value := range envVars {
		cmd.ui.Say("%s: %s", key, terminal.EntityNameColor(value))
	}
}
