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

type UnsetEnv struct {
	ui      terminal.UI
	config  core_config.Reader
	appRepo applications.ApplicationRepository
	appReq  requirements.ApplicationRequirement
}

func init() {
	command_registry.Register(&UnsetEnv{})
}

func (cmd *UnsetEnv) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.appRepo = deps.RepoLocator.GetApplicationRepository()
	return cmd
}

func (cmd *UnsetEnv) MetaData() command_registry.CommandMetadata {
	return command_registry.CommandMetadata{
		Name:        "unset-env",
		Description: T("Remove an env variable"),
		Usage:       T("CF_NAME unset-env APP_NAME ENV_VAR_NAME"),
	}
}

func (cmd *UnsetEnv) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) (reqs []requirements.Requirement, err error) {
	if len(fc.Args()) != 2 {
		cmd.ui.Failed(T("Incorrect Usage. Requires 'app-name env-name' as arguments\n\n") + command_registry.Commands.CommandUsage("unset-env"))
	}

	cmd.appReq = requirementsFactory.NewApplicationRequirement(fc.Args()[0])

	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
		cmd.appReq,
	}
	return
}

func (cmd *UnsetEnv) Execute(c flags.FlagContext) {
	varName := c.Args()[1]
	app := cmd.appReq.GetApplication()

	cmd.ui.Say(T("Removing env variable {{.VarName}} from app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.CurrentUser}}...",
		map[string]interface{}{
			"VarName":     terminal.EntityNameColor(varName),
			"AppName":     terminal.EntityNameColor(app.Name),
			"OrgName":     terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
			"SpaceName":   terminal.EntityNameColor(cmd.config.SpaceFields().Name),
			"CurrentUser": terminal.EntityNameColor(cmd.config.Username())}))

	envParams := app.EnvironmentVars

	if _, ok := envParams[varName]; !ok {
		cmd.ui.Ok()
		cmd.ui.Warn(T("Env variable {{.VarName}} was not set.", map[string]interface{}{"VarName": varName}))
		return
	}

	delete(envParams, varName)

	_, apiErr := cmd.appRepo.Update(app.Guid, models.AppParams{EnvironmentVars: &envParams})
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Ok()
	cmd.ui.Say(T("TIP: Use '{{.Command}}' to ensure your env variable changes take effect",
		map[string]interface{}{"Command": terminal.CommandColor(cf.Name() + " restage")}))
}
