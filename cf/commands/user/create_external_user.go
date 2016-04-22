package user

import (
	"github.com/cloudfoundry/cli/cf"
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/errors"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
)

type CreateExternalUser struct {
	ui       terminal.UI
	config   coreconfig.Reader
	userRepo api.UserRepository
}

func init() {
	commandregistry.Register(&CreateExternalUser{})
}

func (cmd *CreateExternalUser) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "create-external-user",
		Description: T("Create a new user authenticated by an external provider"),
		Usage: []string{
			T("CF_NAME create-external-user USERNAME ORIGIN EXTERNALID"),
		},
	}
}

func (cmd *CreateExternalUser) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	if len(fc.Args()) != 3 {
		usage := commandregistry.Commands.CommandUsage("create-external-user")
		cmd.ui.Failed(T("Incorrect Usage. Requires arguments\n\n") + usage)
	}

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}

	return reqs
}

func (cmd *CreateExternalUser) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.userRepo = deps.RepoLocator.GetUserRepository()
	return cmd
}

func (cmd *CreateExternalUser) Execute(c flags.FlagContext) {
	username := c.Args()[0]
	origin := c.Args()[1]
	externalID := c.Args()[2]

	cmd.ui.Say(T("Creating user {{.TargetUser}}...",
		map[string]interface{}{
			"TargetUser":  terminal.EntityNameColor(username),
			"CurrentUser": terminal.EntityNameColor(cmd.config.Username()),
		}))

	err := cmd.userRepo.CreateExternal(username, origin, externalID)
	switch err.(type) {
	case nil:
	case *errors.ModelAlreadyExistsError:
		cmd.ui.Warn("%s", err.Error())
	default:
		cmd.ui.Failed(T("Error creating user {{.TargetUser}}.\n{{.Error}}",
			map[string]interface{}{
				"TargetUser": terminal.EntityNameColor(username),
				"Error":      err.Error(),
			}))
	}

	cmd.ui.Ok()
	cmd.ui.Say(T("\nTIP: Assign roles with '{{.CurrentUser}} set-org-role' and '{{.CurrentUser}} set-space-role'", map[string]interface{}{"CurrentUser": cf.Name}))
}
