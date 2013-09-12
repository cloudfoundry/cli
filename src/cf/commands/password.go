package commands

import (
	"cf/api"
	"cf/configuration"
	"cf/requirements"
	term "cf/terminal"
	"github.com/codegangsta/cli"
)

type Password struct {
	ui         term.UI
	pwdRepo    api.PasswordRepository
	configRepo configuration.ConfigurationRepository
}

func NewPassword(ui term.UI, pwdRepo api.PasswordRepository, configRepo configuration.ConfigurationRepository) (cmd Password) {
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
	oldPassword := cmd.ui.AskForPassword("Current Password%s", term.PromptColor(">"))
	newPassword := cmd.ui.AskForPassword("New Password%s", term.PromptColor(">"))
	verifiedPassword := cmd.ui.AskForPassword("Verify Password%s", term.PromptColor(">"))

	if verifiedPassword != newPassword {
		cmd.ui.Failed("Password verification does not match")
		return
	}

	score, err := cmd.pwdRepo.GetScore(newPassword)
	if err != nil {
		cmd.ui.Failed(err.Error())
		return
	}
	cmd.ui.Say("Your password strength is: %s", score)

	cmd.ui.Say("Changing password...")
	err = cmd.pwdRepo.UpdatePassword(oldPassword, newPassword)

	if err != nil {
		if err.StatusCode == 401 {
			cmd.ui.Failed("Current password did not match")
		} else {
			cmd.ui.Failed(err.Error())
		}
		return
	}

	cmd.ui.Ok()

	cmd.configRepo.ClearSession()
	cmd.ui.Say("Please log back in.")
}
