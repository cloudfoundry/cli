package serviceauthtoken

import (
	"cf"
	"cf/api"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"fmt"
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

	tokenLabel := c.Args()[0]
	tokenProvider := c.Args()[1]
	token := cf.ServiceAuthToken{
		Label:    tokenLabel,
		Provider: tokenProvider,
	}

	if c.Bool("f") == false {
		response := cmd.ui.Confirm(
			"Are you sure you want to delete %s?%s",
			terminal.EntityNameColor(fmt.Sprintf("%s %s", tokenLabel, tokenProvider)),
			terminal.PromptColor(">"),
		)
		if response == false {
			return
		}
	}

	cmd.ui.Say("Deleting service auth token...")
	token, apiResponse := cmd.authTokenRepo.FindByName(token.FindByNameKey())
	if apiResponse.IsError() {
		cmd.ui.Failed("Error deleting service auth token.\n%s", apiResponse.Message)
		return
	}
	if apiResponse.IsNotFound() {
		cmd.ui.Ok()
		cmd.ui.Warn("Service Auth Token %s %s does not exist.", tokenLabel, tokenProvider)
		return
	}

	apiResponse = cmd.authTokenRepo.Delete(token)
	if apiResponse.IsNotSuccessful() {
		return
	}

	cmd.ui.Ok()
}
