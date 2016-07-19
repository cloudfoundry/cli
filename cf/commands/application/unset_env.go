package application

import (
	"fmt"

	"code.cloudfoundry.org/cli/cf"
	"code.cloudfoundry.org/cli/cf/api/applications"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/flags"
	. "code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"
)

type UnsetEnv struct {
	ui      terminal.UI
	config  coreconfig.Reader
	appRepo applications.Repository
	appReq  requirements.ApplicationRequirement
}

func init() {
	commandregistry.Register(&UnsetEnv{})
}

func (cmd *UnsetEnv) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.appRepo = deps.RepoLocator.GetApplicationRepository()
	return cmd
}

func (cmd *UnsetEnv) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "unset-env",
		Description: T("Remove an env variable"),
		Usage: []string{
			T("CF_NAME unset-env APP_NAME ENV_VAR_NAME"),
		},
	}
}

func (cmd *UnsetEnv) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 2 {
		cmd.ui.Failed(T("Incorrect Usage. Requires 'app-name env-name' as arguments\n\n") + commandregistry.Commands.CommandUsage("unset-env"))
		return nil, fmt.Errorf("Incorrect usage: %d arguments of %d required", len(fc.Args()), 2)
	}

	cmd.appReq = requirementsFactory.NewApplicationRequirement(fc.Args()[0])

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
		cmd.appReq,
	}

	return reqs, nil
}

func (cmd *UnsetEnv) Execute(c flags.FlagContext) error {
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
		return nil
	}

	delete(envParams, varName)

	_, err := cmd.appRepo.Update(app.GUID, models.AppParams{EnvironmentVars: &envParams})
	if err != nil {
		return err
	}

	cmd.ui.Ok()
	cmd.ui.Say(T("TIP: Use '{{.Command}}' to ensure your env variable changes take effect",
		map[string]interface{}{"Command": terminal.CommandColor(cf.Name + " restage")}))
	return nil
}
