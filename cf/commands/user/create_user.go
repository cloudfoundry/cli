package user

import (
	"fmt"

	"code.cloudfoundry.org/cli/cf"
	"code.cloudfoundry.org/cli/cf/api"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/cf/flags"
	. "code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"
)

type CreateUser struct {
	ui       terminal.UI
	config   coreconfig.Reader
	userRepo api.UserRepository
}

func init() {
	commandregistry.Register(&CreateUser{})
}

func (cmd *CreateUser) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "create-user",
		Description: T("Create a new user"),
		Usage: []string{
			T("CF_NAME create-user USERNAME PASSWORD"),
		},
	}
}

func (cmd *CreateUser) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 2 {
		usage := commandregistry.Commands.CommandUsage("create-user")
		cmd.ui.Failed(T("Incorrect Usage. Requires arguments\n\n") + usage)
		return nil, fmt.Errorf("Incorrect usage: %d arguments of %d required", len(fc.Args()), 2)
	}

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}

	return reqs, nil
}

func (cmd *CreateUser) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.userRepo = deps.RepoLocator.GetUserRepository()
	return cmd
}

func (cmd *CreateUser) Execute(c flags.FlagContext) error {
	username := c.Args()[0]
	password := c.Args()[1]

	cmd.ui.Say(T("Creating user {{.TargetUser}}...",
		map[string]interface{}{
			"TargetUser":  terminal.EntityNameColor(username),
			"CurrentUser": terminal.EntityNameColor(cmd.config.Username()),
		}))

	err := cmd.userRepo.Create(username, password)
	switch err.(type) {
	case nil:
	case *errors.ModelAlreadyExistsError:
		cmd.ui.Warn("%s", err.Error())
	default:
		return errors.New(T("Error creating user {{.TargetUser}}.\n{{.Error}}",
			map[string]interface{}{
				"TargetUser": terminal.EntityNameColor(username),
				"Error":      err.Error(),
			}))
	}

	cmd.ui.Ok()
	cmd.ui.Say(T("\nTIP: Assign roles with '{{.CurrentUser}} set-org-role' and '{{.CurrentUser}} set-space-role'", map[string]interface{}{"CurrentUser": cf.Name}))
	return nil
}
