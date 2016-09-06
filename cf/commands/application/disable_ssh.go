package application

import (
	"errors"
	"fmt"

	"code.cloudfoundry.org/cli/cf/api/applications"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/flags"
	. "code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"
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

func (cmd *DisableSSH) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires APP_NAME as argument\n\n") + commandregistry.Commands.CommandUsage("disable-ssh"))
		return nil, fmt.Errorf("Incorrect usage: %d arguments of %d required", len(fc.Args()), 1)
	}

	cmd.appReq = requirementsFactory.NewApplicationRequirement(fc.Args()[0])

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
		cmd.appReq,
	}

	return reqs, nil
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

	cmd.ui.Say(T("Disabling ssh support for '{{.AppName}}'...",
		map[string]interface{}{
			"AppName": app.Name,
		},
	))
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
