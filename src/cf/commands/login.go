package commands

import (
	"cf"
	"cf/api"
	"cf/configuration"
	"cf/errors"
	"cf/models"
	"cf/requirements"
	"cf/terminal"
	"fmt"
	"github.com/codegangsta/cli"
	"strconv"
	"strings"
)

const maxLoginTries = 3
const maxChoices = 50

type Login struct {
	ui                      terminal.UI
	config                  configuration.ReadWriter
	authenticator           api.AuthenticationRepository
	endpointRepo            api.EndpointRepository
	orgRepo                 api.OrganizationRepository
	spaceRepo               api.SpaceRepository
	shouldShowConfiguration bool
}

func NewLogin(ui terminal.UI,
	config configuration.ReadWriter,
	authenticator api.AuthenticationRepository,
	endpointRepo api.EndpointRepository,
	orgRepo api.OrganizationRepository,
	spaceRepo api.SpaceRepository) (cmd Login) {
	return Login{
		ui:                      ui,
		config:                  config,
		authenticator:           authenticator,
		endpointRepo:            endpointRepo,
		orgRepo:                 orgRepo,
		spaceRepo:               spaceRepo,
		shouldShowConfiguration: true,
	}
}

func (cmd Login) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	return
}

func (cmd Login) Run(c *cli.Context) {
	cmd.config.ClearSession()

	defer func() {
		if cmd.shouldShowConfiguration {
			cmd.ui.Say("")
			cmd.ui.ShowConfiguration(cmd.config)
		}
	}()

	(&cmd).setApi(c)
	cmd.authenticate(c)

	orgIsSet := cmd.setOrganization(c)
	if orgIsSet {
		cmd.setSpace(c)
	}
}

func (cmd *Login) setApi(c *cli.Context) {
	api := c.String("a")
	skipSSL := c.Bool("skip-ssl-validation")

	if api == "" {
		api = cmd.config.ApiEndpoint()
		skipSSL = cmd.config.IsSSLDisabled() || skipSSL
	}

	if api == "" {
		api = cmd.ui.Ask("API endpoint%s", terminal.PromptColor(">"))
	} else {
		cmd.ui.Say("API endpoint: %s", terminal.EntityNameColor(api))
	}

	cmd.config.SetSSLDisabled(skipSSL)
	endpoint, err := cmd.endpointRepo.UpdateEndpoint(api)

	if err != nil {
		cmd.config.SetApiEndpoint("")
		cmd.config.SetSSLDisabled(false)

		switch typedErr := err.(type) {
		case *errors.InvalidSSLCert:
			cmd.shouldShowConfiguration = false
			cfLoginCommand := terminal.CommandColor(fmt.Sprintf("%s login --skip-ssl-validation", cf.Name()))
			tipMessage := fmt.Sprintf("TIP: Use '%s' to continue with an insecure API endpoint", cfLoginCommand)
			cmd.ui.Failed("Invalid SSL Cert for %s\n%s", typedErr.URL, tipMessage)
		default:
			cmd.ui.Failed("Invalid API endpoint.\n%s", err.Error())
		}
	}

	if !strings.HasPrefix(endpoint, "https://") {
		cmd.ui.Say(terminal.WarningColor("Warning: Insecure http API endpoint detected: secure https API endpoints are recommended\n"))
	}
}

func (cmd Login) authenticate(c *cli.Context) {
	prompts, err := cmd.authenticator.GetLoginPromptsAndSaveUAAServerURL()
	if err != nil {
		cmd.ui.Failed(err.Error())
	}
	var passwordKey string
	credentials := make(map[string]string)
	for key, prompt := range prompts {
		if prompt.Type == configuration.AuthPromptTypePassword {
			passwordKey = key
		} else if key == "username" && c.String("u") != "" {
			credentials[key] = c.String("u")
		} else {
			credentials[key] = cmd.ui.Ask("%s%s", prompt.DisplayName, terminal.PromptColor(">"))
		}
	}

	password := c.String("p")
	passwordPrompt := prompts[passwordKey]

	for i := 0; i < maxLoginTries; i++ {
		if password == "" || i > 0 {
			password = cmd.ui.AskForPassword("%s%s", passwordPrompt.DisplayName, terminal.PromptColor(">"))
		}

		cmd.ui.Say("Authenticating...")

		credentials[passwordKey] = password
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

		nameString = cmd.ui.Ask("%s%s", itemPrompt, terminal.PromptColor(">"))
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
