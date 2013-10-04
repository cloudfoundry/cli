package commands

import (
	"cf/api"
	"cf/configuration"
	"cf/requirements"
	"cf/terminal"
	"github.com/codegangsta/cli"
)

type Password struct {
	ui         terminal.UI
	pwdRepo    api.PasswordRepository
	configRepo configuration.ConfigurationRepository
}

func NewPassword(ui terminal.UI, pwdRepo api.PasswordRepository, configRepo configuration.ConfigurationRepository) (cmd Password) {
	cmd.ui = ui
	cmd.pwdRepo = pwdRepo
	cmd.configRepo = configRepo
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

	score, apiStatus := cmd.pwdRepo.GetScore(newPassword)
	if apiStatus.NotSuccessful() {
		cmd.ui.Failed(apiStatus.Message)
		return
	}
	cmd.ui.Say("Your password strength is: %s", score)

	cmd.ui.Say("Changing password...")
	apiStatus = cmd.pwdRepo.UpdatePassword(oldPassword, newPassword)

	if apiStatus.NotSuccessful() {
		if apiStatus.StatusCode == 401 {
			cmd.ui.Failed("Current password did not match")
		} else {
			cmd.ui.Failed(apiStatus.Message)
		}
		return
	}

	cmd.ui.Ok()

	cmd.configRepo.ClearSession()
	cmd.ui.Say("Please log in again")
}
