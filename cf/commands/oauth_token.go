package commands

import (
	"github.com/cloudfoundry/cli/cf/api/authentication"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type OAuthToken struct {
	ui       terminal.UI
	config   core_config.ReadWriter
	authRepo authentication.AuthenticationRepository
}

func NewOAuthToken(ui terminal.UI, config core_config.ReadWriter, authRepo authentication.AuthenticationRepository) *OAuthToken {
	return &OAuthToken{
		ui:       ui,
		config:   config,
		authRepo: authRepo,
	}
}

func (cmd *OAuthToken) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}
	return
}

func (cmd *OAuthToken) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "oauth-token",
		Description: T("Retrieve and display the OAuth token for the current session"),
		Usage:       T("CF_NAME oauth-token"),
	}
}

func (cmd *OAuthToken) Run(c *cli.Context) {
	cmd.ui.Say(T("Getting OAuth token..."))

	token, err := cmd.authRepo.RefreshAuthToken()
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	cmd.ui.Ok()
	cmd.ui.Say("")
	cmd.ui.Say(token)
}
