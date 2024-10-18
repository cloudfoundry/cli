package commands

import (
	"code.cloudfoundry.org/cli/v8/cf/api/password"
	"code.cloudfoundry.org/cli/v8/cf/commandregistry"
	"code.cloudfoundry.org/cli/v8/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/v8/cf/errors"
	"code.cloudfoundry.org/cli/v8/cf/flags"
	. "code.cloudfoundry.org/cli/v8/cf/i18n"
	"code.cloudfoundry.org/cli/v8/cf/requirements"
	"code.cloudfoundry.org/cli/v8/cf/terminal"
)

type Password struct {
	ui      terminal.UI
	pwdRepo password.Repository
	config  coreconfig.ReadWriter
}

func init() {
	commandregistry.Register(&Password{})
}

func (cmd *Password) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "passwd",
		ShortName:   "pw",
		Description: T("Change user password"),
		Usage: []string{
			T("CF_NAME passwd"),
		},
	}
}

func (cmd *Password) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}

	return reqs, nil
}

func (cmd *Password) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.pwdRepo = deps.RepoLocator.GetPasswordRepository()
	return cmd
}

func (cmd *Password) Execute(c flags.FlagContext) error {
	oldPassword := cmd.ui.AskForPassword(T("Current Password"))
	newPassword := cmd.ui.AskForPassword(T("New Password"))
	verifiedPassword := cmd.ui.AskForPassword(T("Verify Password"))

	cmd.ui.Say(T("Changing password for user {{.Username}}...", map[string]interface{}{
		"Username": cmd.config.Username(),
	}))

	if verifiedPassword != newPassword {
		return errors.New(T("Password verification does not match"))
	}

	err := cmd.pwdRepo.UpdatePassword(oldPassword, newPassword)

	switch typedErr := err.(type) {
	case nil:
	case errors.HTTPError:
		if typedErr.StatusCode() == 401 {
			return errors.New(T("Current password did not match"))
		}
		return err
	default:
		return err
	}

	cmd.ui.Ok()
	cmd.config.ClearSession()
	cmd.ui.Say(T("Please log in again"))
	return nil
}
