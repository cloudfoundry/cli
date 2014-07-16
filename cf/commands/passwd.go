package commands

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/errors"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type Password struct {
	ui      terminal.UI
	pwdRepo api.PasswordRepository
	config  configuration.ReadWriter
}

func NewPassword(ui terminal.UI, pwdRepo api.PasswordRepository, config configuration.ReadWriter) (cmd Password) {
	cmd.ui = ui
	cmd.pwdRepo = pwdRepo
	cmd.config = config
	return
}

func (cmd Password) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "passwd",
		ShortName:   "pw",
		Description: T("Change user password"),
		Usage:       T("CF_NAME passwd"),
	}
}

func (cmd Password) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	reqs = []requirements.Requirement{requirementsFactory.NewLoginRequirement()}
	return
}

func (cmd Password) Run(c *cli.Context) {
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
