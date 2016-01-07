package application

import (
	"github.com/cloudfoundry/cli/cf/api/applications"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
)

type RenameApp struct {
	ui      terminal.UI
	config  core_config.Reader
	appRepo applications.ApplicationRepository
	appReq  requirements.ApplicationRequirement
}

func init() {
	command_registry.Register(&RenameApp{})
}

func (cmd *RenameApp) MetaData() command_registry.CommandMetadata {
	return command_registry.CommandMetadata{
		Name:        "rename",
		Description: T("Rename an app"),
		Usage:       T("CF_NAME rename APP_NAME NEW_APP_NAME"),
	}
}

func (cmd *RenameApp) Requirements(requirementsFactory requirements.Factory, c flags.FlagContext) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 2 {
		cmd.ui.Failed(T("Incorrect Usage. Requires old app name and new app name as arguments\n\n") + command_registry.Commands.CommandUsage("rename"))
	}

	cmd.appReq = requirementsFactory.NewApplicationRequirement(c.Args()[0])

	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
		cmd.appReq,
	}
	return
}

func (cmd *RenameApp) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.appRepo = deps.RepoLocator.GetApplicationRepository()
	return cmd
}

func (cmd *RenameApp) Execute(c flags.FlagContext) {
	app := cmd.appReq.GetApplication()
	newName := c.Args()[1]

	cmd.ui.Say(T("Renaming app {{.AppName}} to {{.NewName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...",
		map[string]interface{}{
			"AppName":   terminal.EntityNameColor(app.Name),
			"NewName":   terminal.EntityNameColor(newName),
			"OrgName":   terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
			"SpaceName": terminal.EntityNameColor(cmd.config.SpaceFields().Name),
			"Username":  terminal.EntityNameColor(cmd.config.Username())}))

	params := models.AppParams{Name: &newName}

	_, apiErr := cmd.appRepo.Update(app.Guid, params)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}
	cmd.ui.Ok()
}
