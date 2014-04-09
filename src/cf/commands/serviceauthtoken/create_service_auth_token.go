package serviceauthtoken

import (
	"cf/api"
	"cf/configuration"
	"cf/models"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type CreateServiceAuthTokenFields struct {
	ui            terminal.UI
	config        configuration.Reader
	authTokenRepo api.ServiceAuthTokenRepository
}

func NewCreateServiceAuthToken(ui terminal.UI, config configuration.Reader, authTokenRepo api.ServiceAuthTokenRepository) (cmd CreateServiceAuthTokenFields) {
	cmd.ui = ui
	cmd.config = config
	cmd.authTokenRepo = authTokenRepo
	return
}

func (cmd CreateServiceAuthTokenFields) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 3 {
		err = errors.New("Incorrect usage")
		cmd.ui.FailWithUsage(c, "create-service-auth-token")
		return
	}

	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}
	return
}

func (cmd CreateServiceAuthTokenFields) Run(c *cli.Context) {
	cmd.ui.Say("Creating service auth token as %s...", terminal.EntityNameColor(cmd.config.Username()))

	serviceAuthTokenRepo := models.ServiceAuthTokenFields{
		Label:    c.Args()[0],
		Provider: c.Args()[1],
		Token:    c.Args()[2],
	}

	apiErr := cmd.authTokenRepo.Create(serviceAuthTokenRepo)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Ok()
}
