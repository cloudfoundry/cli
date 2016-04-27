package user

import (
	"fmt"
	"strings"

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

type CreateUser struct {
	ui       terminal.UI
	config   coreconfig.Reader
	userRepo api.UserRepository
}

func init() {
	commandregistry.Register(&CreateUser{})
}

func (cmd *CreateUser) MetaData() commandregistry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["origin"] = &flags.StringFlag{Name: "origin", Usage: T("User origin for externally (non-UAA) authenticated users. EXTERNAL_ID must be supplied")}

	return commandregistry.CommandMetadata{
		Name:        "create-user",
		Description: T("Create a new user"),
		Usage: []string{

			fmt.Sprintf("%s:\n", T("Create a UAA authenticated user")),
			"      CF_NAME create-user USERNAME PASSWORD\n\n",

			fmt.Sprintf("  %s:\n", T("Create an externally authenticated user")),
			"      CF_NAME create-user USERNAME EXTERNAL_ID --origin ORIGIN\n\n",
		},
		Examples: []string{
			"CF_NAME create-user j.smith@mydomain.com S3cr3t                                                  # UAA-authenticated user",
			"CF_NAME create-user j.smith@mydomain.com 'cn=1234,ou=MYCORD,dc=Dev,dc=HQ,dc=MYCO' --origin ldap  # LDAP-authenticated user",
		},
		Flags: fs,
	}
}

func (cmd *CreateUser) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	if len(fc.Args()) != 2 {
		usage := commandregistry.Commands.CommandUsage("create-user")
		cmd.ui.Failed(T("Incorrect Usage. Requires arguments\n\n") + usage)
	}

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}

	return reqs
}

func (cmd *CreateUser) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.userRepo = deps.RepoLocator.GetUserRepository()
	return cmd
}

func (cmd *CreateUser) Execute(c flags.FlagContext) {

	username := c.Args()[0]

	var err error

	cmd.ui.Say(T("Creating user {{.TargetUser}}...",
		map[string]interface{}{
			"TargetUser":  terminal.EntityNameColor(username),
			"CurrentUser": terminal.EntityNameColor(cmd.config.Username()),
		}))

	if !c.IsSet("origin") || c.IsSet("origin") && strings.ToLower(c.String("origin")) == "uaa" {
		password := c.Args()[1]
		err = cmd.userRepo.Create(username, password)
	} else {
		externalid := c.Args()[1]
		origin := strings.ToLower(c.String("origin"))
		err = cmd.userRepo.CreateExternal(username, origin, externalid)
	}

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
