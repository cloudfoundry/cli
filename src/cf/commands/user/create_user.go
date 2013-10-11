package user

import (
	"cf"
	"cf/api"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type CreateUser struct {
	ui       terminal.UI
	userRepo api.UserRepository
}

func NewCreateUser(ui terminal.UI, userRepo api.UserRepository) (cmd CreateUser) {
	cmd.ui = ui
	cmd.userRepo = userRepo
	return
}

func (cmd CreateUser) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 2 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "create-user")
	}

	reqs = append(reqs, reqFactory.NewLoginRequirement())

	return
}

func (cmd CreateUser) Run(c *cli.Context) {
	username := c.Args()[0]
	password := c.Args()[1]

	cmd.ui.Say("Creating user %s...", username)

	user := cf.User{
		Username: username,
		Password: password,
	}
	apiResponse := cmd.userRepo.Create(user)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed("Error creating user %s.\n%s", terminal.EntityNameColor(username), apiResponse.Message)
		return
	}

	cmd.ui.Ok()

	cmd.ui.Say("\nTIP: Assign roles with '%s set-org-role' and '%s set-space-role'", cf.Name, cf.Name)
}
