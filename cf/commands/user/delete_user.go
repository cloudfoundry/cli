package user

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

type DeleteUser struct {
	ui       terminal.UI
	config   configuration.Reader
	userRepo api.UserRepository
}

func NewDeleteUser(ui terminal.UI, config configuration.Reader, userRepo api.UserRepository) (cmd DeleteUser) {
	cmd.ui = ui
	cmd.config = config
	cmd.userRepo = userRepo
	return
}

func (cmd DeleteUser) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "delete-user",
		Description: T("Delete a user"),
		Usage:       T("CF_NAME delete-user USERNAME [-f]"),
		Flags: []cli.Flag{
			cli.BoolFlag{Name: "f", Usage: T("Force deletion without confirmation")},
		},
	}
}

func (cmd DeleteUser) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 1 {
		err = errors.New(T("Invalid usage"))
		cmd.ui.FailWithUsage(c)
		return
	}

	reqs = append(reqs, requirementsFactory.NewLoginRequirement())

	return
}

func (cmd DeleteUser) Run(c *cli.Context) {
	username := c.Args()[0]
	force := c.Bool("f")

	if !force && !cmd.ui.ConfirmDelete(T("user"), username) {
		return
	}

	cmd.ui.Say(T("Deleting user {{.TargetUser}} as {{.CurrentUser}}...",
		map[string]interface{}{
			"TargetUser":  terminal.EntityNameColor(username),
			"CurrentUser": terminal.EntityNameColor(cmd.config.Username()),
		}))

	user, apiErr := cmd.userRepo.FindByUsername(username)
	switch apiErr.(type) {
	case nil:
	case *errors.ModelNotFoundError:
		cmd.ui.Ok()
		cmd.ui.Warn(T("User {{.TargetUser}} does not exist.", map[string]interface{}{"TargetUser": username}))
		return
	default:
		cmd.ui.Failed(apiErr.Error())
		return
	}

	apiErr = cmd.userRepo.Delete(user.Guid)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Ok()
}
