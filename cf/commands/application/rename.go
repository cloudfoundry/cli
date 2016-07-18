package application

import (
	"github.com/cloudfoundry/cli/cf/api/applications"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/flags"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
)

type RenameApp struct {
	ui      terminal.UI
	config  coreconfig.Reader
	appRepo applications.Repository
	appReq  requirements.ApplicationRequirement
}

func init() {
	commandregistry.Register(&RenameApp{})
}

func (cmd *RenameApp) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "rename",
		Description: T("Rename an app"),
		Usage: []string{
			T("CF_NAME rename APP_NAME NEW_APP_NAME"),
		},
	}
}

func (cmd *RenameApp) Requirements(requirementsFactory requirements.Factory, c flags.FlagContext) []requirements.Requirement {
	if len(c.Args()) != 2 {
		cmd.ui.Failed(T("Incorrect Usage. Requires old app name and new app name as arguments\n\n") + commandregistry.Commands.CommandUsage("rename"))
	}

	cmd.appReq = requirementsFactory.NewApplicationRequirement(c.Args()[0])

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
		cmd.appReq,
	}

	return reqs
}

func (cmd *RenameApp) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.appRepo = deps.RepoLocator.GetApplicationRepository()
	return cmd
}

func (cmd *RenameApp) Execute(c flags.FlagContext) error {
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

	_, err := cmd.appRepo.Update(app.GUID, params)
	if err != nil {
		return err
	}
	cmd.ui.Ok()
	return err
}
