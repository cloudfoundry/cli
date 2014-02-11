package commands

import (
	"cf/api"
	"cf/configuration"
	"cf/requirements"
	"cf/terminal"
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

func (cmd Password) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	reqs = []requirements.Requirement{
		reqFactory.NewValidAccessTokenRequirement(),
	}
	return
}

func (cmd Password) Run(c *cli.Context) {
	oldPassword := cmd.ui.AskForPassword("Current Password%s", terminal.PromptColor(">"))
	newPassword := cmd.ui.AskForPassword("New Password%s", terminal.PromptColor(">"))
	verifiedPassword := cmd.ui.AskForPassword("Verify Password%s", terminal.PromptColor(">"))

	if verifiedPassword != newPassword {
		cmd.ui.Failed("Password verification does not match")
		return
	}

	cmd.ui.Say("Changing password...")
	apiResponse := cmd.pwdRepo.UpdatePassword(oldPassword, newPassword)

	if apiResponse.IsNotSuccessful() {
		if apiResponse.StatusCode == 401 {
			cmd.ui.Failed("Current password did not match")
		} else {
			cmd.ui.Failed(apiResponse.Message)
		}
		return
	}

	cmd.ui.Ok()

	cmd.config.ClearSession()
	cmd.ui.Say("Please log in again")
}
