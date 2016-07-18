package application

import (
	"errors"
	"fmt"

	"github.com/cloudfoundry/cli/cf/api/applications"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/flags"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
)

type DisableSSH struct {
	ui      terminal.UI
	config  coreconfig.Reader
	appReq  requirements.ApplicationRequirement
	appRepo applications.Repository
}

func init() {
	commandregistry.Register(&DisableSSH{})
}

func (cmd *DisableSSH) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "disable-ssh",
		Description: T("Disable ssh for the application"),
		Usage: []string{
			T("CF_NAME disable-ssh APP_NAME"),
		},
	}
}

func (cmd *DisableSSH) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires APP_NAME as argument\n\n") + commandregistry.Commands.CommandUsage("disable-ssh"))
	}

	cmd.appReq = requirementsFactory.NewApplicationRequirement(fc.Args()[0])

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
		cmd.appReq,
	}

	return reqs
}

func (cmd *DisableSSH) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.appRepo = deps.RepoLocator.GetApplicationRepository()
	return cmd
}

func (cmd *DisableSSH) Execute(fc flags.FlagContext) error {
	app := cmd.appReq.GetApplication()

	if !app.EnableSSH {
		cmd.ui.Say(fmt.Sprintf(T("ssh support is already disabled")+" for '%s'", app.Name))
		return nil
	}

	cmd.ui.Say(fmt.Sprintf(T("Disabling ssh support for '%s'..."), app.Name))
	cmd.ui.Say("")

	enable := false
	updatedApp, err := cmd.appRepo.Update(app.GUID, models.AppParams{EnableSSH: &enable})
	if err != nil {
		return errors.New(T("Error disabling ssh support for ") + app.Name + ": " + err.Error())
	}

	if !updatedApp.EnableSSH {
		cmd.ui.Ok()
	} else {
		return errors.New(T("ssh support is not disabled for ") + app.Name)
	}
	return nil
}
