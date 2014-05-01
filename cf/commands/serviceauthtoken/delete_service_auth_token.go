package serviceauthtoken

import (
	"fmt"
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type DeleteServiceAuthTokenFields struct {
	ui            terminal.UI
	config        configuration.Reader
	authTokenRepo api.ServiceAuthTokenRepository
}

func NewDeleteServiceAuthToken(ui terminal.UI, config configuration.Reader, authTokenRepo api.ServiceAuthTokenRepository) (cmd DeleteServiceAuthTokenFields) {
	cmd.ui = ui
	cmd.config = config
	cmd.authTokenRepo = authTokenRepo
	return
}

func (command DeleteServiceAuthTokenFields) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "delete-service-auth-token",
		Description: "Delete a service auth token",
		Usage:       "CF_NAME delete-service-auth-token LABEL PROVIDER [-f]",
		Flags: []cli.Flag{
			cli.BoolFlag{Name: "f", Usage: "Force deletion without confirmation"},
		},
	}
}

func (cmd DeleteServiceAuthTokenFields) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 2 {
		err = errors.New("Incorrect usage")
		cmd.ui.FailWithUsage(c, "delete-service-auth-token")
		return
	}

	reqs = append(reqs, requirementsFactory.NewLoginRequirement())
	return
}

func (cmd DeleteServiceAuthTokenFields) Run(c *cli.Context) {
	tokenLabel := c.Args()[0]
	tokenProvider := c.Args()[1]

	if c.Bool("f") == false {
		if !cmd.ui.ConfirmDelete("service auth token", fmt.Sprintf("%s %s", tokenLabel, tokenProvider)) {
			return
		}
	}

	cmd.ui.Say("Deleting service auth token as %s", terminal.EntityNameColor(cmd.config.Username()))
	token, apiErr := cmd.authTokenRepo.FindByLabelAndProvider(tokenLabel, tokenProvider)

	switch apiErr.(type) {
	case nil:
	case *errors.ModelNotFoundError:
		cmd.ui.Ok()
		cmd.ui.Warn("Service Auth Token %s %s does not exist.", tokenLabel, tokenProvider)
		return
	default:
		cmd.ui.Failed(apiErr.Error())
	}

	apiErr = cmd.authTokenRepo.Delete(token)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Ok()
}
