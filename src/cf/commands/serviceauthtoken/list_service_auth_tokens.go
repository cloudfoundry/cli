package serviceauthtoken

import (
	"cf/api"
	"cf/requirements"
	"cf/terminal"
	"github.com/codegangsta/cli"
)

type ListServiceAuthTokens struct {
	ui            terminal.UI
	authTokenRepo api.ServiceAuthTokenRepository
}

func NewListServiceAuthTokens(ui terminal.UI, authTokenRepo api.ServiceAuthTokenRepository) (cmd ListServiceAuthTokens) {
	cmd.ui = ui
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
	cmd.ui.Say("Getting service auth tokens...")
	authTokens, apiResponse := cmd.authTokenRepo.FindAll()
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}
	cmd.ui.Ok()
	cmd.ui.Say("")

	table := [][]string{
		{"Label", "Provider", "Guid"},
	}

	for _, authToken := range authTokens {
		table = append(table, []string{authToken.Label, authToken.Provider, authToken.Guid})
	}

	cmd.ui.DisplayTable(table)
}
