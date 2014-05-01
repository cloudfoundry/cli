package serviceauthtoken

import (
	"errors"
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type UpdateServiceAuthTokenFields struct {
	ui            terminal.UI
	config        configuration.Reader
	authTokenRepo api.ServiceAuthTokenRepository
}

func NewUpdateServiceAuthToken(ui terminal.UI, config configuration.Reader, authTokenRepo api.ServiceAuthTokenRepository) (cmd UpdateServiceAuthTokenFields) {
	cmd.ui = ui
	cmd.config = config
	cmd.authTokenRepo = authTokenRepo
	return
}

func (command UpdateServiceAuthTokenFields) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "update-service-auth-token",
		Description: "Update a service auth token",
		Usage:       "CF_NAME update-service-auth-token LABEL PROVIDER TOKEN",
	}
}

func (cmd UpdateServiceAuthTokenFields) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 3 {
		err = errors.New("Incorrect usage")
		cmd.ui.FailWithUsage(c, "update-service-auth-token")
		return
	}

	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}
	return
}

func (cmd UpdateServiceAuthTokenFields) Run(c *cli.Context) {
	cmd.ui.Say("Updating service auth token as %s...", terminal.EntityNameColor(cmd.config.Username()))

	serviceAuthToken, apiErr := cmd.authTokenRepo.FindByLabelAndProvider(c.Args()[0], c.Args()[1])
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	serviceAuthToken.Token = c.Args()[2]

	apiErr = cmd.authTokenRepo.Update(serviceAuthToken)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Ok()
}
