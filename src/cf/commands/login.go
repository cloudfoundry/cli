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
	prompt := terminal.PromptColor(">")

	api := c.String("a")
	if api == "" {
		api = cmd.ui.Ask("API endpoint%s", prompt)
	}

	apiResponse = cmd.endpointRepo.UpdateEndpoint(api)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	username := c.String("u")
	if username == "" {
		username = cmd.ui.Ask("Username%s", prompt)
	}

	password := c.String("p")

	for i := 0; i < maxLoginTries; i++ {
		if password == "" || i > 0 {
			password = cmd.ui.AskForPassword("Password%s", terminal.PromptColor(">"))
		}

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

	orgName := c.String("o")
	if orgName == "" {
		orgName = cmd.ui.Ask("Org%s", prompt)
	}

	organization, apiResponse := cmd.orgRepo.FindByName(orgName)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed("Error finding org %s\n%s", terminal.EntityNameColor(orgName), apiResponse.Message)
		return
	}

	err := cmd.configRepo.SetOrganization(organization)
	if err != nil {
		cmd.ui.Failed("Error setting org %s in config file\n%s", terminal.EntityNameColor(orgName), err.Error())
		return
	}

	spaceName := c.String("s")
	if spaceName == "" {
		spaceName = cmd.ui.Ask("Space%s", prompt)
	}

	space, apiResponse := cmd.spaceRepo.FindByName(spaceName)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed("Error finding space %s\n%s", terminal.EntityNameColor(spaceName), apiResponse.Message)
		return
	}

	err = cmd.configRepo.SetSpace(space)
	if err != nil {
		cmd.ui.Failed("Error setting space %s in config file\n%s", terminal.EntityNameColor(spaceName), err.Error())
		return
	}

	cmd.ui.ShowConfiguration(cmd.config)
	return
}
