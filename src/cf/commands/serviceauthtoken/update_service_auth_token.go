package serviceauthtoken

import (
	"cf/api"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type UpdateServiceAuthToken struct {
	ui            terminal.UI
	authTokenRepo api.ServiceAuthTokenRepository
}

func NewUpdateServiceAuthToken(ui terminal.UI, authTokenRepo api.ServiceAuthTokenRepository) (cmd UpdateServiceAuthToken) {
	cmd.ui = ui
	cmd.authTokenRepo = authTokenRepo
	return
}

func (cmd UpdateServiceAuthToken) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 3 {
		err = errors.New("Incorrect usage")
		cmd.ui.FailWithUsage(c, "update-service-auth-token")
		return
	}

	reqs = []requirements.Requirement{
		reqFactory.NewLoginRequirement(),
	}
	return
}

func (cmd UpdateServiceAuthToken) Run(c *cli.Context) {
	cmd.ui.Say("Updating service auth token...")

	serviceAuthToken, apiResponse := cmd.authTokenRepo.FindByLabelAndProvider(c.Args()[0], c.Args()[1])
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	serviceAuthToken.Token = c.Args()[2]

	apiResponse = cmd.authTokenRepo.Update(serviceAuthToken)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Ok()
}
