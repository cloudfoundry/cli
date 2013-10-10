package serviceauthtoken

import (
	"cf"
	"cf/api"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type DeleteServiceAuthToken struct {
	ui            terminal.UI
	authTokenRepo api.ServiceAuthTokenRepository
}

func NewDeleteServiceAuthToken(ui terminal.UI, authTokenRepo api.ServiceAuthTokenRepository) (cmd DeleteServiceAuthToken) {
	cmd.ui = ui
	cmd.authTokenRepo = authTokenRepo
	return
}

func (cmd DeleteServiceAuthToken) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 2 {
		err = errors.New("Incorrect usage")
		cmd.ui.FailWithUsage(c, "delete-service-auth-token")
		return
	}

	reqs = append(reqs, reqFactory.NewLoginRequirement())
	return
}

func (cmd DeleteServiceAuthToken) Run(c *cli.Context) {
	cmd.ui.Say("Deleting service auth token...")

	tokenLabel := c.Args()[0]
	tokenProvider := c.Args()[1]
	token := cf.ServiceAuthToken{
		Label:    tokenLabel,
		Provider: tokenProvider,
	}

	token, apiResponse := cmd.authTokenRepo.FindByName(token.FindByNameKey())
	if apiResponse.IsNotSuccessful() {
		return
	}

	apiResponse = cmd.authTokenRepo.Delete(token)
	if apiResponse.IsNotSuccessful() {
		return
	}

	cmd.ui.Ok()
}
