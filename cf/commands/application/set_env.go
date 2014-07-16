package application

import (
	"github.com/cloudfoundry/cli/cf"
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type SetEnv struct {
	ui      terminal.UI
	config  configuration.Reader
	appRepo api.ApplicationRepository
	appReq  requirements.ApplicationRequirement
}

func NewSetEnv(ui terminal.UI, config configuration.Reader, appRepo api.ApplicationRepository) *SetEnv {
	return &SetEnv{
		ui:      ui,
		config:  config,
		appRepo: appRepo,
	}
}

func (cmd *SetEnv) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:            "set-env",
		ShortName:       "se",
		Description:     T("Set an env variable for an app"),
		Usage:           T("CF_NAME set-env APP NAME VALUE"),
		SkipFlagParsing: true,
	}
}

func (cmd *SetEnv) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 3 {
		cmd.ui.FailWithUsage(c)
	}

	cmd.appReq = requirementsFactory.NewApplicationRequirement(c.Args()[0])
	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
		cmd.appReq,
	}
	return
}

func (cmd *SetEnv) Run(c *cli.Context) {
	varName := c.Args()[1]
	varValue := c.Args()[2]
	app := cmd.appReq.GetApplication()

	cmd.ui.Say(T("Setting env variable '{{.VarName}}' to '{{.VarValue}}' for app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.CurrentUser}}...",
		map[string]interface{}{
			"VarName":     terminal.EntityNameColor(varName),
			"VarValue":    terminal.EntityNameColor(varValue),
			"AppName":     terminal.EntityNameColor(app.Name),
			"OrgName":     terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
			"SpaceName":   terminal.EntityNameColor(cmd.config.SpaceFields().Name),
			"CurrentUser": terminal.EntityNameColor(cmd.config.Username())}))

	if len(app.EnvironmentVars) == 0 {
		app.EnvironmentVars = map[string]string{}
	}
	envParams := app.EnvironmentVars
	envParams[varName] = varValue

	_, apiErr := cmd.appRepo.Update(app.Guid, models.AppParams{EnvironmentVars: &envParams})

	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Ok()
	cmd.ui.Say(T("TIP: Use '{{.Command}}' to ensure your env variable changes take effect",
		map[string]interface{}{"Command": terminal.CommandColor(cf.Name() + " restage")}))
}
