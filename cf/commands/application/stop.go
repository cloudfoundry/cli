package application

import (
	"github.com/cloudfoundry/cli/cf/flags"
	. "github.com/cloudfoundry/cli/cf/i18n"

	"github.com/cloudfoundry/cli/cf/api/applications"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
)

//go:generate counterfeiter . Stopper

type Stopper interface {
	commandregistry.Command
	ApplicationStop(app models.Application, orgName string, spaceName string) (updatedApp models.Application, err error)
}

type Stop struct {
	ui      terminal.UI
	config  coreconfig.Reader
	appRepo applications.Repository
	appReq  requirements.ApplicationRequirement
}

func init() {
	commandregistry.Register(&Stop{})
}

func (cmd *Stop) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "stop",
		ShortName:   "sp",
		Description: T("Stop an app"),
		Usage: []string{
			T("CF_NAME stop APP_NAME"),
		},
	}
}

func (cmd *Stop) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + commandregistry.Commands.CommandUsage("stop"))
	}

	cmd.appReq = requirementsFactory.NewApplicationRequirement(fc.Args()[0])

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
		cmd.appReq,
	}

	return reqs
}

func (cmd *Stop) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.appRepo = deps.RepoLocator.GetApplicationRepository()
	return cmd
}

func (cmd *Stop) ApplicationStop(app models.Application, orgName, spaceName string) (models.Application, error) {
	var updatedApp models.Application

	if app.State == models.ApplicationStateStopped {
		updatedApp = app
		return updatedApp, nil
	}

	cmd.ui.Say(T("Stopping app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.CurrentUser}}...",
		map[string]interface{}{
			"AppName":     terminal.EntityNameColor(app.Name),
			"OrgName":     terminal.EntityNameColor(orgName),
			"SpaceName":   terminal.EntityNameColor(spaceName),
			"CurrentUser": terminal.EntityNameColor(cmd.config.Username())}))

	state := "STOPPED"
	updatedApp, err := cmd.appRepo.Update(app.GUID, models.AppParams{State: &state})
	if err != nil {
		return models.Application{}, err
	}

	cmd.ui.Ok()
	return updatedApp, nil
}

func (cmd *Stop) Execute(c flags.FlagContext) error {
	app := cmd.appReq.GetApplication()
	if app.State == models.ApplicationStateStopped {
		cmd.ui.Say(terminal.WarningColor(T("App ") + app.Name + T(" is already stopped")))
	} else {
		_, err := cmd.ApplicationStop(app, cmd.config.OrganizationFields().Name, cmd.config.SpaceFields().Name)
		if err != nil {
			return err
		}
	}
	return nil
}
