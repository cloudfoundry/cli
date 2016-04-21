package commands

import (
	"github.com/cloudfoundry/cli/cf/api/password"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/errors"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
)

type Password struct {
	ui      terminal.UI
	pwdRepo password.PasswordRepository
	config  coreconfig.ReadWriter
}

func init() {
	commandregistry.Register(&Password{})
}

func (cmd *Password) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "passwd",
		ShortName:   "pw",
		Description: T("Change user password"),
		Usage: []string{
			T("CF_NAME passwd"),
		},
	}
}

func (cmd *Password) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}

	return reqs
}

func (cmd *Password) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
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
	case errors.HTTPError:
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
