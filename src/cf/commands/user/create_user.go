package user

import (
	"cf"
	"cf/api"
	"cf/configuration"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type CreateUserFields struct {
	ui       terminal.UI
	config   configuration.Reader
	userRepo api.UserRepository
}

func NewCreateUser(ui terminal.UI, config configuration.Reader, userRepo api.UserRepository) (cmd CreateUserFields) {
	cmd.ui = ui
	cmd.config = config
	cmd.userRepo = userRepo
	return
}

func (cmd CreateUserFields) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 2 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "create-user")
	}

	reqs = append(reqs, reqFactory.NewLoginRequirement())

	return
}

func (cmd CreateUserFields) Run(c *cli.Context) {
	username := c.Args()[0]
	password := c.Args()[1]

	cmd.ui.Say("Creating user %s as %s...",
		terminal.EntityNameColor(username),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	apiErr := cmd.userRepo.Create(username, password)
	if apiErr != nil {
		cmd.ui.Failed("Error creating user %s.\n%s", terminal.EntityNameColor(username), apiErr.Error())
		return
	}

	cmd.ui.Ok()

	cmd.ui.Say("\nTIP: Assign roles with '%s set-org-role' and '%s set-space-role'", cf.Name(), cf.Name())
}
