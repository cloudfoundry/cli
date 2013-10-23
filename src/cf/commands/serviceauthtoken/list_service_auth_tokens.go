package serviceauthtoken

import (
	"cf/api"
	"cf/configuration"
	"cf/requirements"
	"cf/terminal"
	"github.com/codegangsta/cli"
)

type ListServiceAuthTokens struct {
	ui            terminal.UI
	config        *configuration.Configuration
	authTokenRepo api.ServiceAuthTokenRepository
}

func NewListServiceAuthTokens(ui terminal.UI, config *configuration.Configuration, authTokenRepo api.ServiceAuthTokenRepository) (cmd ListServiceAuthTokens) {
	cmd.ui = ui
	cmd.config = config
	cmd.authTokenRepo = authTokenRepo
	return
}

func (cmd ListServiceAuthTokens) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	reqs = []requirements.Requirement{
		reqFactory.NewLoginRequirement(),
	}
	return
}

func (cmd ListServiceAuthTokens) Run(c *cli.Context) {
	cmd.ui.Say("Getting service auth tokens as %s...", terminal.EntityNameColor(cmd.config.Username()))
	authTokens, apiResponse := cmd.authTokenRepo.FindAll()
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}
	cmd.ui.Ok()
	cmd.ui.Say("")

	table := [][]string{
		{"label", "provider"},
	}

	for _, authToken := range authTokens {
		table = append(table, []string{authToken.Label, authToken.Provider})
	}

	cmd.ui.DisplayTable(table)
}
