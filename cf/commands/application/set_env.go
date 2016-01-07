package application

import (
	"github.com/cloudfoundry/cli/cf"
	"github.com/cloudfoundry/cli/cf/api/applications"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
)

type SetEnv struct {
	ui      terminal.UI
	config  core_config.Reader
	appRepo applications.ApplicationRepository
	appReq  requirements.ApplicationRequirement
}

func init() {
	command_registry.Register(&SetEnv{})
}

func (cmd *SetEnv) MetaData() command_registry.CommandMetadata {
	return command_registry.CommandMetadata{
		Name:            "set-env",
		ShortName:       "se",
		Description:     T("Set an env variable for an app"),
		Usage:           T("CF_NAME set-env APP_NAME ENV_VAR_NAME ENV_VAR_VALUE"),
		SkipFlagParsing: true,
	}
}

func (cmd *SetEnv) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) (reqs []requirements.Requirement, err error) {
	if len(fc.Args()) != 3 {
		cmd.ui.Failed(T("Incorrect Usage. Requires 'app-name env-name env-value' as arguments\n\n") + command_registry.Commands.CommandUsage("set-env"))
	}

	cmd.appReq = requirementsFactory.NewApplicationRequirement(fc.Args()[0])

	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
		cmd.appReq,
	}
	return
}

func (cmd *SetEnv) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.appRepo = deps.RepoLocator.GetApplicationRepository()
	return cmd
}

func (cmd *SetEnv) Execute(c flags.FlagContext) {
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
		app.EnvironmentVars = map[string]interface{}{}
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
