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

type CreateLDAPUser struct {
	ui       terminal.UI
	config   coreconfig.Reader
	userRepo api.UserRepository
}

func init() {
	commandregistry.Register(&CreateLDAPUser{})
}

func (cmd *CreateLDAPUser) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "create-ldap-user",
		Description: T("Create a new user authenticated by LDAP"),
		Usage: []string{
			T("CF_NAME create-ldap-user USERNAME EXTERNALID"),
		},
	}
}

func (cmd *CreateLDAPUser) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	if len(fc.Args()) != 2 {
		usage := commandregistry.Commands.CommandUsage("create-ldap-user")
		cmd.ui.Failed(T("Incorrect Usage. Requires arguments\n\n") + usage)
	}

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}

	return reqs
}

func (cmd *CreateLDAPUser) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.userRepo = deps.RepoLocator.GetUserRepository()
	return cmd
}

func (cmd *CreateLDAPUser) Execute(c flags.FlagContext) {
	username := c.Args()[0]
	externalID := c.Args()[1]

	cmd.ui.Say(T("Creating user {{.TargetUser}}...",
		map[string]interface{}{
			"TargetUser":  terminal.EntityNameColor(username),
			"CurrentUser": terminal.EntityNameColor(cmd.config.Username()),
		}))

	err := cmd.userRepo.CreateLDAP(username, externalID)
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
