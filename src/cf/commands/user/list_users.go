package user

import (
	"cf/api"
	"cf/requirements"
	"cf/terminal"
	"github.com/codegangsta/cli"
)

type ListUsers struct {
	ui       terminal.UI
	userRepo api.UserRepository
}

func NewListUsers(ui terminal.UI, userRepo api.UserRepository) (cmd ListUsers) {
	cmd.ui = ui
	cmd.userRepo = userRepo
	return
}

func (cmd ListUsers) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	reqs = append(reqs, reqFactory.NewLoginRequirement())
	return
}

func (cmd ListUsers) Run(c *cli.Context) {
	cmd.ui.Say("Getting users in all orgs and spaces...")

	users, apiResponse := cmd.userRepo.FindAll()
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Ok()
	cmd.ui.Say("")

	for _, user := range users {
		if user.IsAdmin {
			cmd.ui.Say("%s %s", user.Username, terminal.EntityNameColor("(admin)"))
		} else {
			cmd.ui.Say("%s", user.Username)
		}
	}
}
