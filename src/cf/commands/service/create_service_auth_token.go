package service

import (
	"cf"
	"cf/api"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type CreateServiceAuthToken struct {
	ui            terminal.UI
	authTokenRepo api.ServiceAuthTokenRepository
}

func NewCreateServiceAuthToken(ui terminal.UI, authTokenRepo api.ServiceAuthTokenRepository) (cmd CreateServiceAuthToken) {
	cmd.ui = ui
	cmd.authTokenRepo = authTokenRepo
	return
}

func (cmd CreateServiceAuthToken) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 2 {
		err = errors.New("Incorrect usage")
		cmd.ui.FailWithUsage(c, "create-service-auth-token")
		return
	}

	reqs = []requirements.Requirement{
		reqFactory.NewLoginRequirement(),
	}
	return
}

func (cmd CreateServiceAuthToken) Run(c *cli.Context) {
	cmd.ui.Say("Creating service auth token...")

	serviceAuthTokenRepo := cf.ServiceAuthToken{
		Label:    c.Args()[0],
		Token:    c.Args()[1],
		Provider: c.String("p"),
	}

	apiResponse := cmd.authTokenRepo.Create(serviceAuthTokenRepo)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Ok()
}
