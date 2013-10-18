package commands

import (
	"cf/api"
	"cf/configuration"
	"cf/net"
	"cf/requirements"
	"cf/terminal"
	"github.com/codegangsta/cli"
)

const maxLoginTries = 3

type Login struct {
	ui            terminal.UI
	config        *configuration.Configuration
	configRepo    configuration.ConfigurationRepository
	authenticator api.AuthenticationRepository
	endpointRepo  api.EndpointRepository
	orgRepo       api.OrganizationRepository
	spaceRepo     api.SpaceRepository
}

func NewLogin(ui terminal.UI,
	configRepo configuration.ConfigurationRepository,
	authenticator api.AuthenticationRepository,
	endpointRepo api.EndpointRepository,
	orgRepo api.OrganizationRepository,
	spaceRepo api.SpaceRepository) (cmd Login) {

	cmd.ui = ui
	cmd.configRepo = configRepo
	cmd.config, _ = configRepo.Get()
	cmd.authenticator = authenticator
	cmd.endpointRepo = endpointRepo
	cmd.orgRepo = orgRepo
	cmd.spaceRepo = spaceRepo

	return
}

func (cmd Login) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	return
}

func (cmd Login) Run(c *cli.Context) {
	var apiResponse net.ApiResponse

	api := c.String("a")
	username := c.String("u")
	password := c.String("p")
	orgName := c.String("o")
	spaceName := c.String("s")

	prompt := terminal.PromptColor(">")

	if api == "" {
		api = cmd.ui.Ask("API endpoint%s", prompt)
	}

	apiResponse = cmd.endpointRepo.UpdateEndpoint(api)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	if username == "" {
		username = cmd.ui.Ask("Username%s", prompt)
	}

	if password == "" {
		for i := 0; i < maxLoginTries; i++ {
			password = cmd.ui.AskForPassword("Password%s", terminal.PromptColor(">"))
			cmd.ui.Say("Authenticating...")

			apiResponse = cmd.authenticator.Authenticate(username, password)
			if apiResponse.IsNotSuccessful() {
				cmd.ui.Say(apiResponse.Message)
				continue
			}
			break
		}

		if apiResponse.IsNotSuccessful() {
			cmd.ui.Failed("Unable to authenticate.")
			return
		}

		cmd.ui.Ok()
	}

	if orgName == "" {
		orgName = cmd.ui.Ask("Org%s", prompt)
	}

	organization, apiResponse := cmd.orgRepo.FindByName(orgName)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed("Error finding org %s\n%s", terminal.EntityNameColor(orgName), apiResponse.Message)
	}

	err := cmd.configRepo.SetOrganization(organization)
	if err != nil {
		cmd.ui.Failed("Error setting org %s in config file\n%s", terminal.EntityNameColor(orgName), err.Error())
	}

	if spaceName == "" {
		spaceName = cmd.ui.Ask("Space%s", prompt)
	}

	space, apiResponse := cmd.spaceRepo.FindByName(spaceName)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed("Error finding space %s\n%s", terminal.EntityNameColor(spaceName), apiResponse.Message)
	}

	err = cmd.configRepo.SetSpace(space)
	if err != nil {
		cmd.ui.Failed("Error setting space %s in config file\n%s", terminal.EntityNameColor(spaceName), err.Error())
	}

	cmd.ui.ShowConfiguration(cmd.config)
	return
}
