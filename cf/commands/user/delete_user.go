package user

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/flags"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
)

type DeleteUser struct {
	ui       terminal.UI
	config   coreconfig.Reader
	userRepo api.UserRepository
}

func init() {
	commandregistry.Register(&DeleteUser{})
}

func (cmd *DeleteUser) MetaData() commandregistry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["f"] = &flags.BoolFlag{ShortName: "f", Usage: T("Force deletion without confirmation")}

	return commandregistry.CommandMetadata{
		Name:        "delete-user",
		Description: T("Delete a user"),
		Usage: []string{
			T("CF_NAME delete-user USERNAME [-f]"),
		},
		Flags: fs,
	}
}

func (cmd *DeleteUser) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + commandregistry.Commands.CommandUsage("delete-user"))
	}

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}

	return reqs
}

func (cmd *DeleteUser) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.userRepo = deps.RepoLocator.GetUserRepository()
	return cmd
}

func (cmd *DeleteUser) Execute(c flags.FlagContext) error {
	username := c.Args()[0]
	force := c.Bool("f")

	if !force && !cmd.ui.ConfirmDelete(T("user"), username) {
		return nil
	}

	cmd.ui.Say(T("Deleting user {{.TargetUser}} as {{.CurrentUser}}...",
		map[string]interface{}{
			"TargetUser":  terminal.EntityNameColor(username),
			"CurrentUser": terminal.EntityNameColor(cmd.config.Username()),
		}))

	user, err := cmd.userRepo.FindByUsername(username)
	switch err.(type) {
	case nil:
	case *errors.ModelNotFoundError:
		cmd.ui.Ok()
		cmd.ui.Warn(T("User {{.TargetUser}} does not exist.", map[string]interface{}{"TargetUser": username}))
		return nil
	default:
		return err
	}

	err = cmd.userRepo.Delete(user.GUID)
	if err != nil {
		return err
	}

	cmd.ui.Ok()
	return nil
}
