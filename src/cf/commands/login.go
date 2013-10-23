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
	apiResponse := cmd.setApi(c)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	apiResponse = cmd.authenticate(c)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed("Unable to authenticate.")
		return
	}

	apiResponse = cmd.setOrganization(c)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	apiResponse = cmd.setSpace(c)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.ShowConfiguration(cmd.config)
	return
}

func (cmd Login) setApi(c *cli.Context) (apiResponse net.ApiResponse) {
	api := c.String("a")
	if api == "" {
		api = cmd.config.Target
	}
	if api == "" {
		api = cmd.ui.Ask("API endpoint%s", terminal.PromptColor(">"))
	}

	apiResponse = cmd.endpointRepo.UpdateEndpoint(api)
	return
}

func (cmd Login) authenticate(c *cli.Context) (apiResponse net.ApiResponse) {
	username := c.String("u")
	if username == "" {
		username = cmd.ui.Ask("Username%s", terminal.PromptColor(">"))
	}

	password := c.String("p")

	for i := 0; i < maxLoginTries; i++ {
		if password == "" || i > 0 {
			password = cmd.ui.AskForPassword("Password%s", terminal.PromptColor(">"))
		}

		cmd.ui.Say("Authenticating...")

		apiResponse = cmd.authenticator.Authenticate(username, password)
		if apiResponse.IsSuccessful() {
			break
		}

		cmd.ui.Say(apiResponse.Message)
	}
	return
}

func (cmd Login) setOrganization(c *cli.Context) (apiResponse net.ApiResponse) {
	orgName := c.String("o")

	// Prompt for org name
	if orgName == "" {
		orgName = cmd.ui.Ask("Org%s", terminal.PromptColor(">"))
	}

	// Find org
	organization, apiResponse := cmd.orgRepo.FindByName(orgName)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed("Error finding org %s\n%s", terminal.EntityNameColor(orgName), apiResponse.Message)
		return
	}

	// Target org
	err := cmd.configRepo.SetOrganization(organization)
	if err != nil {
		apiResponse = net.NewApiResponseWithMessage("Error setting org %s in config file\n%s", terminal.EntityNameColor(orgName), err.Error())
	}
	return
}

func (cmd Login) setSpace(c *cli.Context) (apiResponse net.ApiResponse) {
	spaceName := c.String("s")

	// Prompt for space name
	if spaceName == "" {
		spaceName = cmd.ui.Ask("Space%s", terminal.PromptColor(">"))
	}

	// Find space
	space, apiResponse := cmd.spaceRepo.FindByName(spaceName)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed("Error finding space %s\n%s", terminal.EntityNameColor(spaceName), apiResponse.Message)
		return
	}

	// Target space
	err := cmd.configRepo.SetSpace(space)
	if err != nil {
		apiResponse = net.NewApiResponseWithMessage("Error setting space %s in config file\n%s", terminal.EntityNameColor(spaceName), err.Error())
		return
	}
	return
}
