package commands

import (
	"cf/api"
	"cf/requirements"
	term "cf/terminal"
	"github.com/codegangsta/cli"
)

type Password struct {
	ui      term.UI
	pwdRepo api.PasswordRepository
}

func NewPassword(ui term.UI, pwdRepo api.PasswordRepository) (cmd Password) {
	cmd.ui = ui
	cmd.pwdRepo = pwdRepo
	return
}

func (cmd Password) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	reqs = []requirements.Requirement{
		reqFactory.NewLoginRequirement(),
	}
	return
}

func (cmd Password) Run(c *cli.Context) {
	oldPassword := cmd.ui.AskForPassword("Current Password%s", term.Cyan(">"))
	newPassword := cmd.ui.AskForPassword("New Password%s", term.Cyan(">"))
	cmd.ui.AskForPassword("Verify Password%s", term.Cyan(">"))

	score, err := cmd.pwdRepo.GetScore(newPassword)
	if err != nil {
		cmd.ui.Failed("Error scoring password", err)
		return
	}
	cmd.ui.Say("Your password strength is: %s", score)

	cmd.ui.Say("Changing password...")
	err = cmd.pwdRepo.UpdatePassword(oldPassword, newPassword)

	if err != nil {
		cmd.ui.Failed("Error changing password", err)
		return
	}

	cmd.ui.Ok()
}
