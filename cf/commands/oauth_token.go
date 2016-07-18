package commands

import (
	"github.com/cloudfoundry/cli/cf/api/authentication"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/flags"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/plugin/models"
)

type OAuthToken struct {
	ui          terminal.UI
	config      coreconfig.ReadWriter
	authRepo    authentication.Repository
	pluginModel *plugin_models.GetOauthToken_Model
	pluginCall  bool
}

func init() {
	commandregistry.Register(&OAuthToken{})
}

func (cmd *OAuthToken) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "oauth-token",
		Description: T("Retrieve and display the OAuth token for the current session"),
		Usage: []string{
			T("CF_NAME oauth-token"),
		},
	}
}

func (cmd *OAuthToken) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}

	return reqs
}

func (cmd *OAuthToken) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.authRepo = deps.RepoLocator.GetAuthenticationRepository()
	cmd.pluginCall = pluginCall
	cmd.pluginModel = deps.PluginModels.OauthToken
	return cmd
}

func (cmd *OAuthToken) Execute(c flags.FlagContext) error {
	token, err := cmd.authRepo.RefreshAuthToken()
	if err != nil {
		return err
	}

	if cmd.pluginCall {
		cmd.pluginModel.Token = token
	} else {
		cmd.ui.Say(token)
	}
	return nil
}
