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
	config        configuration.Reader
	authTokenRepo api.ServiceAuthTokenRepository
}

func NewListServiceAuthTokens(ui terminal.UI, config configuration.Reader, authTokenRepo api.ServiceAuthTokenRepository) (cmd ListServiceAuthTokens) {
	cmd.ui = ui
	cmd.config = config
	cmd.authTokenRepo = authTokenRepo
	return
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
	rows := [][]string{}

	for _, authToken := range authTokens {
		rows = append(rows, []string{authToken.Label, authToken.Provider})
	}

	table.Print(rows)
}
