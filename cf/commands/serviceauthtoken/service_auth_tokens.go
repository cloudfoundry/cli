package serviceauthtoken

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type ListServiceAuthTokens struct {
	ui            terminal.UI
	config        configuration.Reader
	authTokenRepo api.ServiceAuthTokenRepository
}

func NewListServiceAuthTokens(ui terminal.UI, config configuration.Reader, authTokenRepo api.ServiceAuthTokenRepository) (cmd ListServiceAuthTokens) {
	cmd.ui = ui
	cmd.config = config
	cmd.authTokenRepo = authTokenRepo
	return
}

func (command ListServiceAuthTokens) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "service-auth-tokens",
		Description: "List service auth tokens",
		Usage:       "CF_NAME service-auth-tokens",
	}
}

func (cmd ListServiceAuthTokens) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}
	return
}

func (cmd ListServiceAuthTokens) Run(c *cli.Context) {
	cmd.ui.Say("Getting service auth tokens as %s...", terminal.EntityNameColor(cmd.config.Username()))
	authTokens, apiErr := cmd.authTokenRepo.FindAll()
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}
	cmd.ui.Ok()
	cmd.ui.Say("")

	table := terminal.NewTable(cmd.ui, []string{"label", "provider"})

	for _, authToken := range authTokens {
		table.Add([]string{authToken.Label, authToken.Provider})
	}

	table.Print()
}
