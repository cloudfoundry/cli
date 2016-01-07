package application

import (
	"github.com/cloudfoundry/cli/cf/api/applications"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
)

type Restage struct {
	ui                terminal.UI
	config            core_config.Reader
	appRepo           applications.ApplicationRepository
	appStagingWatcher ApplicationStagingWatcher
}

func init() {
	command_registry.Register(&Restage{})
}

func (cmd *Restage) MetaData() command_registry.CommandMetadata {
	return command_registry.CommandMetadata{
		Name:        "restage",
		ShortName:   "rg",
		Description: T("Restage an app"),
		Usage:       T("CF_NAME restage APP_NAME"),
	}
}

func (cmd *Restage) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) (reqs []requirements.Requirement, err error) {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + command_registry.Commands.CommandUsage("restage"))
	}

	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
	}
	return
}

func (cmd *Restage) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.appRepo = deps.RepoLocator.GetApplicationRepository()

	//get command from registry for dependency
	commandDep := command_registry.Commands.FindCommand("start")
	commandDep = commandDep.SetDependency(deps, false)
	cmd.appStagingWatcher = commandDep.(ApplicationStagingWatcher)

	return cmd
}

func (cmd *Restage) Execute(c flags.FlagContext) {
	app, err := cmd.appRepo.Read(c.Args()[0])
	if notFound, ok := err.(*errors.ModelNotFoundError); ok {
		cmd.ui.Failed(notFound.Error())
	}

	cmd.ui.Say(T("Restaging app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.CurrentUser}}...",
		map[string]interface{}{
			"AppName":     terminal.EntityNameColor(app.Name),
			"OrgName":     terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
			"SpaceName":   terminal.EntityNameColor(cmd.config.SpaceFields().Name),
			"CurrentUser": terminal.EntityNameColor(cmd.config.Username()),
		}))

	app.PackageState = ""

	cmd.appStagingWatcher.ApplicationWatchStaging(app, cmd.config.OrganizationFields().Name, cmd.config.SpaceFields().Name, func(app models.Application) (models.Application, error) {
		return app, cmd.appRepo.CreateRestageRequest(app.Guid)
	})
}
