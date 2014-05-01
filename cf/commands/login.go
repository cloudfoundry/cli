package commands

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/flag_helpers"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
	"strconv"
)

const maxLoginTries = 3
const maxChoices = 50

type Login struct {
	ui            terminal.UI
	config        configuration.ReadWriter
	authenticator api.AuthenticationRepository
	endpointRepo  api.EndpointRepository
	orgRepo       api.OrganizationRepository
	spaceRepo     api.SpaceRepository
}

func NewLogin(ui terminal.UI,
	config configuration.ReadWriter,
	authenticator api.AuthenticationRepository,
	endpointRepo api.EndpointRepository,
	orgRepo api.OrganizationRepository,
	spaceRepo api.SpaceRepository) (cmd Login) {
	return Login{
		ui:            ui,
		config:        config,
		authenticator: authenticator,
		endpointRepo:  endpointRepo,
		orgRepo:       orgRepo,
		spaceRepo:     spaceRepo,
	}
}

func (command Login) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "login",
		ShortName:   "l",
		Description: "Log user in",
		Usage: "CF_NAME login [-a API_URL] [-u USERNAME] [-p PASSWORD] [-o ORG] [-s SPACE]\n\n" +
			terminal.WarningColor("WARNING:\n   Providing your password as a command line option is highly discouraged\n   Your password may be visible to others and may be recorded in your shell history\n\n") +
			"EXAMPLE:\n" +
			"   CF_NAME login (omit username and password to login interactively -- CF_NAME will prompt for both)\n" +
			"   CF_NAME login -u name@example.com -p pa55woRD (specify username and password as arguments)\n" +
			"   CF_NAME login -u name@example.com -p \"my password\" (use quotes for passwords with a space)\n" +
			"   CF_NAME login -u name@example.com -p \"\\\"password\\\"\" (escape quotes if used in password)" +
			"   CF_NAME login --sso (CF_NAME will provide a url to obtain a one-time password to login)",
		Flags: []cli.Flag{
			flag_helpers.NewStringFlag("a", "API endpoint (e.g. https://api.example.com)"),
			flag_helpers.NewStringFlag("u", "Username"),
			flag_helpers.NewStringFlag("p", "Password"),
			flag_helpers.NewStringFlag("o", "Org"),
			flag_helpers.NewStringFlag("s", "Space"),
			cli.BoolFlag{Name: "sso", Usage: "Use a one-time password to login"},
			cli.BoolFlag{Name: "skip-ssl-validation", Usage: "Please don't"},
		},
	}
}

func (cmd Login) GetRequirements(_ requirements.Factory, _ *cli.Context) (reqs []requirements.Requirement, err error) {
	return
}

func (cmd Login) Run(c *cli.Context) {
	cmd.config.ClearSession()

	endpoint, skipSSL := cmd.decideEndpoint(c)
	NewApi(cmd.ui, cmd.config, cmd.endpointRepo).setApiEndpoint(endpoint, skipSSL)

	defer func() {
		cmd.ui.Say("")
		cmd.ui.ShowConfiguration(cmd.config)
	}()

	// We thought we would never need to explicitly branch in this code
	// for anything as simple as authentication, but it turns out that our
	// assumptions did not match reality.

	// When SAML is enabled (but not configured) then the UAA/Login server
	// will always returns password prompts that includes the Passcode field.
	// Users can authenticate with:
	//   EITHER   username and password
	//   OR       a one-time passcode

	if c.Bool("sso") {
		cmd.authenticateSSO(c)
	} else {
		cmd.authenticate(c)
	}

	orgIsSet := cmd.setOrganization(c)

	if orgIsSet {
		cmd.setSpace(c)
	}
}

func (cmd Login) decideEndpoint(c *cli.Context) (string, bool) {
	endpoint := c.String("a")
	skipSSL := c.Bool("skip-ssl-validation")
	if endpoint == "" {
		endpoint = cmd.config.ApiEndpoint()
		skipSSL = cmd.config.IsSSLDisabled() || skipSSL
	}

	if endpoint == "" {
		endpoint = cmd.ui.Ask("API endpoint")
	} else {
		cmd.ui.Say("API endpoint: %s", terminal.EntityNameColor(endpoint))
	}

	return endpoint, skipSSL
}

func (cmd Login) authenticateSSO(c *cli.Context) {
	prompts, err := cmd.authenticator.GetLoginPromptsAndSaveUAAServerURL()
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	credentials := make(map[string]string)
	passcode := prompts["passcode"]

	for i := 0; i < maxLoginTries; i++ {
		credentials["passcode"] = cmd.ui.AskForPassword("%s", passcode.DisplayName)

		cmd.ui.Say("Authenticating...")
		err = cmd.authenticator.Authenticate(credentials)

		if err == nil {
			cmd.ui.Ok()
			cmd.ui.Say("")
			break
		}

		cmd.ui.Say(err.Error())
	}

	if err != nil {
		cmd.ui.Failed("Unable to authenticate.")
	}
}

func (cmd Login) authenticate(c *cli.Context) {
	usernameFlagValue := c.String("u")
	passwordFlagValue := c.String("p")

	prompts, err := cmd.authenticator.GetLoginPromptsAndSaveUAAServerURL()
	if err != nil {
		cmd.ui.Failed(err.Error())
	}
	passwordKeys := []string{}
	credentials := make(map[string]string)
	for key, prompt := range prompts {
		if prompt.Type == configuration.AuthPromptTypePassword {
			if key == "passcode" {
				continue
			}

			passwordKeys = append(passwordKeys, key)
		} else if key == "username" && usernameFlagValue != "" {
			credentials[key] = usernameFlagValue
		} else {
			credentials[key] = cmd.ui.Ask("%s", prompt.DisplayName)
		}
	}

	for i := 0; i < maxLoginTries; i++ {
		for _, key := range passwordKeys {
			if key == "password" && passwordFlagValue != "" {
				credentials[key] = passwordFlagValue
				passwordFlagValue = ""
			} else {
				credentials[key] = cmd.ui.AskForPassword("%s", prompts[key].DisplayName)
			}
		}

		cmd.ui.Say("Authenticating...")
		err = cmd.authenticator.Authenticate(credentials)

		if err == nil {
			cmd.ui.Ok()
			cmd.ui.Say("")
			break
		}

		cmd.ui.Say(err.Error())
	}

	if err != nil {
		cmd.ui.Failed("Unable to authenticate.")
	}
}

func (cmd Login) setOrganization(c *cli.Context) (isOrgSet bool) {
	orgName := c.String("o")

	if orgName == "" {
		availableOrgs := []models.Organization{}
		apiErr := cmd.orgRepo.ListOrgs(func(o models.Organization) bool {
			availableOrgs = append(availableOrgs, o)
			return len(availableOrgs) < maxChoices
		})
		if apiErr != nil {
			cmd.ui.Failed("Error finding avilable orgs\n%s", apiErr.Error())
		}

		if len(availableOrgs) == 1 {
			cmd.targetOrganization(availableOrgs[0])
			return true
		}

		orgName = cmd.promptForOrgName(availableOrgs)
		if orgName == "" {
			cmd.ui.Say("")
			return false
		}
	}

	org, err := cmd.orgRepo.FindByName(orgName)
	if err != nil {
		cmd.ui.Failed("Error finding org %s\n%s", terminal.EntityNameColor(orgName), err.Error())
	}

	cmd.targetOrganization(org)
	return true
}

func (cmd Login) promptForOrgName(orgs []models.Organization) string {
	orgNames := []string{}
	for _, org := range orgs {
		orgNames = append(orgNames, org.Name)
	}

	return cmd.promptForName(orgNames, "Select an org (or press enter to skip):", "Org")
}

func (cmd Login) targetOrganization(org models.Organization) {
	cmd.config.SetOrganizationFields(org.OrganizationFields)
	cmd.ui.Say("Targeted org %s\n", terminal.EntityNameColor(org.Name))
}

func (cmd Login) setSpace(c *cli.Context) {
	spaceName := c.String("s")

	if spaceName == "" {
		var availableSpaces []models.Space
		err := cmd.spaceRepo.ListSpaces(func(space models.Space) bool {
			availableSpaces = append(availableSpaces, space)
			return (len(availableSpaces) < maxChoices)
		})
		if err != nil {
			cmd.ui.Failed("Error finding available spaces\n%s", err.Error())
		}

		// Target only space if possible
		if len(availableSpaces) == 1 {
			cmd.targetSpace(availableSpaces[0])
			return
		}

		spaceName = cmd.promptForSpaceName(availableSpaces)
		if spaceName == "" {
			cmd.ui.Say("")
			return
		}
	}

	space, err := cmd.spaceRepo.FindByName(spaceName)
	if err != nil {
		cmd.ui.Failed("Error finding space %s\n%s", terminal.EntityNameColor(spaceName), err.Error())
	}

	cmd.targetSpace(space)
}

func (cmd Login) promptForSpaceName(spaces []models.Space) string {
	spaceNames := []string{}
	for _, space := range spaces {
		spaceNames = append(spaceNames, space.Name)
	}

	return cmd.promptForName(spaceNames, "Select a space (or press enter to skip):", "Space")
}

func (cmd Login) targetSpace(space models.Space) {
	cmd.config.SetSpaceFields(space.SpaceFields)
	cmd.ui.Say("Targeted space %s\n", terminal.EntityNameColor(space.Name))
}

func (cmd Login) promptForName(names []string, listPrompt, itemPrompt string) string {
	nameIndex := 0
	var nameString string
	for nameIndex < 1 || nameIndex > len(names) {
		var err error

		// list header
		cmd.ui.Say(listPrompt)

		// only display list if it is shorter than maxChoices
		if len(names) < maxChoices {
			for i, name := range names {
				cmd.ui.Say("%d. %s", i+1, name)
			}
		} else {
			cmd.ui.Say("There are too many options to display, please type in the name.")
		}

		nameString = cmd.ui.Ask("%s", itemPrompt)
		if nameString == "" {
			return ""
		}

		nameIndex, err = strconv.Atoi(nameString)

		if err != nil {
			nameIndex = 1
			return nameString
		}
	}

	return names[nameIndex-1]
}
