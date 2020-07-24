package application

import (
	"fmt"

	"code.cloudfoundry.org/cli/cf/api/applications"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/cf/flags"
	. "code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"
)

type Restage struct {
	ui                terminal.UI
	config            coreconfig.Reader
	appRepo           applications.Repository
	appStagingWatcher StagingWatcher
}

func init() {
	commandregistry.Register(&Restage{})
}

func (cmd *Restage) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "restage",
		ShortName:   "rg",
		Description: T("Restage an app"),
		Usage: []string{
			T("CF_NAME restage APP_NAME"),
		},
	}
}

func (cmd *Restage) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + commandregistry.Commands.CommandUsage("restage"))
		return nil, fmt.Errorf("Incorrect usage: %d arguments of %d required", len(fc.Args()), 1)
	}

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
	}

	return reqs, nil
}

func (cmd *Restage) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.appRepo = deps.RepoLocator.GetApplicationRepository()

	//get command from registry for dependency
	commandDep := commandregistry.Commands.FindCommand("start")
	commandDep = commandDep.SetDependency(deps, false)
	cmd.appStagingWatcher = commandDep.(StagingWatcher)

	return cmd
}

func (cmd *Restage) Execute(c flags.FlagContext) error {
	app, err := cmd.appRepo.Read(c.Args()[0])
	if notFound, ok := err.(*errors.ModelNotFoundError); ok {
		return notFound
	}

	cmd.ui.Say(T("Restaging app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.CurrentUser}}...",
		map[string]interface{}{
			"AppName":     terminal.EntityNameColor(app.Name),
			"OrgName":     terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
			"SpaceName":   terminal.EntityNameColor(cmd.config.SpaceFields().Name),
			"CurrentUser": terminal.EntityNameColor(cmd.config.Username()),
		}))

	app.PackageState = ""

	_, err = cmd.appStagingWatcher.WatchStaging(app, cmd.config.OrganizationFields().Name, cmd.config.SpaceFields().Name, func(app models.Application) (models.Application, error) {
		return app, cmd.appRepo.CreateRestageRequest(app.GUID)
	})
	if err != nil {
		cmd.ui.Say(T("Failed to watch staging of app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.CurrentUser}}...",
			map[string]interface{}{
				"AppName":     terminal.EntityNameColor(app.Name),
				"OrgName":     terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
				"SpaceName":   terminal.EntityNameColor(cmd.config.SpaceFields().Name),
				"CurrentUser": terminal.EntityNameColor(cmd.config.Username()),
			}))
	}
	return nil
}
