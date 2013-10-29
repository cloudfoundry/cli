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
	"strings"
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
	oldUserName := cmd.config.Username()

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

	userChanged := (cmd.config.Username() != oldUserName && oldUserName != "")

	apiResponse = cmd.setOrganization(c, userChanged)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	apiResponse = cmd.setSpace(c, userChanged)
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

	endpoint, apiResponse := cmd.endpointRepo.UpdateEndpoint(api)
	if !strings.HasPrefix(endpoint, "https://") {
		cmd.ui.Say(terminal.WarningColor("Warning: Insecure http API endpoint detected: secure https API endpoints are recommended\n"))
	}

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

func (cmd Login) setOrganization(c *cli.Context, userChanged bool) (apiResponse net.ApiResponse) {
	orgName := c.String("o")

	if orgName == "" {
		// If the user is changing, clear out the org
		if userChanged {
			cmd.config.Organization = cf.Organization{}
		}

		// Reuse org in config
		if cmd.config.HasOrganization() && !userChanged {
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
	orgNames := []string{}
	for _, org := range orgs {
		orgNames = append(orgNames, org.Name)
	}

	return cmd.promptForName(orgNames, "Select an org:", "Org")
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

func (cmd Login) setSpace(c *cli.Context, userChanged bool) (apiResponse net.ApiResponse) {
	spaceName := c.String("s")

	if spaceName == "" {
		// If user is changing, clear the space
		if userChanged {
			cmd.config.Space = cf.Space{}
		}
		// Reuse space in config
		if cmd.config.HasSpace() && !userChanged {
			return
		}

		// Get available spaces
		var availableSpaces []cf.Space

		availableSpaces, apiResponse = cmd.spaceRepo.FindAll()
		if apiResponse.IsNotSuccessful() {
			cmd.ui.Failed("Error finding avilable spaces\n%s", apiResponse.Message)
			return
		}

		// Target only space if possible
		if len(availableSpaces) == 1 {
			return cmd.targetSpace(availableSpaces[0])
		}

		spaceName = cmd.promptForSpaceName(availableSpaces)
	}

	// Find space
	space, apiResponse := cmd.spaceRepo.FindByName(spaceName)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed("Error finding space %s\n%s", terminal.EntityNameColor(spaceName), apiResponse.Message)
		return
	}

	return cmd.targetSpace(space)
}

func (cmd Login) promptForSpaceName(spaces []cf.Space) string {
	spaceNames := []string{}
	for _, space := range spaces {
		spaceNames = append(spaceNames, space.Name)
	}

	return cmd.promptForName(spaceNames, "Select a space:", "Space")
}

func (cmd Login) targetSpace(space cf.Space) (apiResponse net.ApiResponse) {
	err := cmd.configRepo.SetSpace(space)
	if err != nil {
		apiResponse = net.NewApiResponseWithMessage("Error setting space %s in config file\n%s",
			terminal.EntityNameColor(space.Name),
			err.Error(),
		)
	}
	return
}

func (cmd Login) promptForName(names []string, listPrompt, itemPrompt string) string {
	nameIndex := 0
	var nameString string
	for nameIndex < 1 || nameIndex > len(names) {
		var err error

		// list header
		cmd.ui.Say(listPrompt)

		// only display list if it is shorter than 50
		if len(names) < 50 {
			for i, name := range names {
				cmd.ui.Say("%d. %s", i+1, name)
			}
		}

		nameString = cmd.ui.Ask("%s%s", itemPrompt, terminal.PromptColor(">"))
		nameIndex, err = strconv.Atoi(nameString)

		if err != nil {
			cmd.ui.Say("")
			nameIndex = 1
			return nameString
		}
	}

	return names[nameIndex-1]
}
