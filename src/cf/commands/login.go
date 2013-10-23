package commands

import (
	"cf"
	"cf/api"
	"cf/configuration"
	"cf/net"
	"cf/requirements"
	"cf/terminal"
	"github.com/codegangsta/cli"
	"strconv"
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
			cmd.ui.Ok()
			cmd.ui.Say("")
			break
		}

		cmd.ui.Say(apiResponse.Message)
	}
	return
}

func (cmd Login) setOrganization(c *cli.Context) (apiResponse net.ApiResponse) {
	orgName := c.String("o")

	if orgName == "" {
		// Reuse org in config
		if cmd.config.HasOrganization() {
			return
		}

		// Get available orgs
		var availableOrgs []cf.Organization

		availableOrgs, apiResponse = cmd.orgRepo.FindAll()
		if apiResponse.IsNotSuccessful() {
			cmd.ui.Failed("Error finding avilable orgs\n%s", apiResponse.Message)
			return
		}

		// Target only org if possible
		if len(availableOrgs) == 1 {
			return cmd.targetOrganization(availableOrgs[0])
		}

		orgName = cmd.promptForOrgName(availableOrgs)
	}

	// Find org
	org, apiResponse := cmd.orgRepo.FindByName(orgName)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed("Error finding org %s\n%s", terminal.EntityNameColor(orgName), apiResponse.Message)
		return
	}

	return cmd.targetOrganization(org)
}

func (cmd Login) promptForOrgName(orgs []cf.Organization) string {
	orgIndex := 0

	for orgIndex < 1 || orgIndex > len(orgs) {
		var err error

		cmd.ui.Say("Select an org:")
		for i, o := range orgs {
			cmd.ui.Say("%d. %s", i+1, o.Name)
		}
		orgNumber := cmd.ui.Ask("Org%s", terminal.PromptColor(">"))
		orgIndex, err = strconv.Atoi(orgNumber)

		if err != nil {
			orgIndex = 0
			cmd.ui.Say("")
		}
	}

	return orgs[orgIndex-1].Name
}

func (cmd Login) targetOrganization(org cf.Organization) (apiResponse net.ApiResponse) {
	err := cmd.configRepo.SetOrganization(org)

	if err != nil {
		apiResponse = net.NewApiResponseWithMessage("Error setting org %s in config file\n%s",
			terminal.EntityNameColor(org.Name),
			err.Error(),
		)
	}
	return
}

func (cmd Login) setSpace(c *cli.Context) (apiResponse net.ApiResponse) {
	spaceName := c.String("s")

	// Reuse space in config
	if spaceName == "" && cmd.config.HasSpace() {
		return
	}

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
