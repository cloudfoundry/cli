package commands

import (
	"github.com/cloudfoundry/cli/cf/api/authentication"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
)

type OAuthToken struct {
	ui       terminal.UI
	config   core_config.ReadWriter
	authRepo authentication.AuthenticationRepository
}

func init() {
	command_registry.Register(&OAuthToken{})
}

func (cmd *OAuthToken) MetaData() command_registry.CommandMetadata {
	return command_registry.CommandMetadata{
		Name:        "oauth-token",
		Description: T("Retrieve and display the OAuth token for the current session"),
		Usage:       T("CF_NAME oauth-token"),
	}
}

func (cmd *OAuthToken) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) (reqs []requirements.Requirement, err error) {
	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}
	return
}

func (cmd *OAuthToken) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.authRepo = deps.RepoLocator.GetAuthenticationRepository()
	return cmd
}

func (cmd *OAuthToken) Execute(c flags.FlagContext) {
	cmd.ui.Say(T("Getting OAuth token..."))

	token, err := cmd.authRepo.RefreshAuthToken()
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	cmd.ui.Ok()
	cmd.ui.Say("")
	cmd.ui.Say(token)
}
