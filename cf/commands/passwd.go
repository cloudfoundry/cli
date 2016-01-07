package commands

import (
	"github.com/cloudfoundry/cli/cf/api/password"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
)

type Password struct {
	ui      terminal.UI
	pwdRepo password.PasswordRepository
	config  core_config.ReadWriter
}

func init() {
	command_registry.Register(&Password{})
}

func (cmd *Password) MetaData() command_registry.CommandMetadata {
	return command_registry.CommandMetadata{
		Name:        "passwd",
		ShortName:   "pw",
		Description: T("Change user password"),
		Usage:       T("CF_NAME passwd"),
	}
}

func (cmd *Password) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) (reqs []requirements.Requirement, err error) {
	reqs = []requirements.Requirement{requirementsFactory.NewLoginRequirement()}
	return
}

func (cmd *Password) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.pwdRepo = deps.RepoLocator.GetPasswordRepository()
	return cmd
}

func (cmd *Password) Execute(c flags.FlagContext) {
	oldPassword := cmd.ui.AskForPassword(T("Current Password"))
	newPassword := cmd.ui.AskForPassword(T("New Password"))
	verifiedPassword := cmd.ui.AskForPassword(T("Verify Password"))

	if verifiedPassword != newPassword {
		cmd.ui.Failed(T("Password verification does not match"))
		return
	}

	cmd.ui.Say(T("Changing password..."))
	apiErr := cmd.pwdRepo.UpdatePassword(oldPassword, newPassword)

	switch typedErr := apiErr.(type) {
	case nil:
	case errors.HttpError:
		if typedErr.StatusCode() == 401 {
			cmd.ui.Failed(T("Current password did not match"))
		} else {
			cmd.ui.Failed(apiErr.Error())
		}
	default:
		cmd.ui.Failed(apiErr.Error())
	}

	cmd.ui.Ok()
	cmd.config.ClearSession()
	cmd.ui.Say(T("Please log in again"))
}
