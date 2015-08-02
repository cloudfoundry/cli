package commands

import (
	"strconv"

	"github.com/cloudfoundry/cli/cf/command_registry"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/flags"
	"github.com/cloudfoundry/cli/flags/flag"

	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/api/authentication"
	"github.com/cloudfoundry/cli/cf/api/organizations"
	"github.com/cloudfoundry/cli/cf/api/spaces"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
)

const maxLoginTries = 3
const maxChoices = 50

type Login struct {
	ui            terminal.UI
	config        core_config.ReadWriter
	authenticator authentication.AuthenticationRepository
	endpointRepo  api.EndpointRepository
	orgRepo       organizations.OrganizationRepository
	spaceRepo     spaces.SpaceRepository
}

func init() {
	command_registry.Register(&Login{})
}

func (cmd *Login) MetaData() command_registry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["a"] = &cliFlags.StringFlag{Name: "a", Usage: T("API endpoint (e.g. https://api.example.com)")}
	fs["u"] = &cliFlags.StringFlag{Name: "u", Usage: T("Username")}
	fs["p"] = &cliFlags.StringFlag{Name: "p", Usage: T("Password")}
	fs["o"] = &cliFlags.StringFlag{Name: "o", Usage: T("Org")}
	fs["s"] = &cliFlags.StringFlag{Name: "s", Usage: T("Space")}
	fs["sso"] = &cliFlags.BoolFlag{Name: "sso", Usage: T("Use a one-time password to login")}
	fs["skip-ssl-validation"] = &cliFlags.BoolFlag{Name: "skip-ssl-validation", Usage: T("Please don't")}

	return command_registry.CommandMetadata{
		Name:        "login",
		ShortName:   "l",
		Description: T("Log user in"),
		Usage: T("CF_NAME login [-a API_URL] [-u USERNAME] [-p PASSWORD] [-o ORG] [-s SPACE]\n\n") +
			terminal.WarningColor(T("WARNING:\n   Providing your password as a command line option is highly discouraged\n   Your password may be visible to others and may be recorded in your shell history\n\n")) + T("EXAMPLE:\n") + T("   CF_NAME login (omit username and password to login interactively -- CF_NAME will prompt for both)\n") + T("   CF_NAME login -u name@example.com -p pa55woRD (specify username and password as arguments)\n") + T("   CF_NAME login -u name@example.com -p \"my password\" (use quotes for passwords with a space)\n") + T("   CF_NAME login -u name@example.com -p \"\\\"password\\\"\" (escape quotes if used in password)\n") + T("   CF_NAME login --sso (CF_NAME will provide a url to obtain a one-time password to login)"),
		Flags: fs,
	}
}

func (cmd *Login) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) (reqs []requirements.Requirement, err error) {
	return
}

func (cmd *Login) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.authenticator = deps.RepoLocator.GetAuthenticationRepository()
	cmd.endpointRepo = deps.RepoLocator.GetEndpointRepository()
	cmd.orgRepo = deps.RepoLocator.GetOrganizationRepository()
	cmd.spaceRepo = deps.RepoLocator.GetSpaceRepository()
	return cmd
}

func (cmd *Login) Execute(c flags.FlagContext) {
	cmd.config.ClearSession()

	endpoint, skipSSL := cmd.decideEndpoint(c)

	Api{
		ui:           cmd.ui,
		config:       cmd.config,
		endpointRepo: cmd.endpointRepo,
	}.setApiEndpoint(endpoint, skipSSL, cmd.MetaData().Name)

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
	cmd.ui.NotifyUpdateIfNeeded(cmd.config)
}

func (cmd Login) decideEndpoint(c flags.FlagContext) (string, bool) {
	endpoint := c.String("a")
	skipSSL := c.Bool("skip-ssl-validation")
	if endpoint == "" {
		endpoint = cmd.config.ApiEndpoint()
		skipSSL = cmd.config.IsSSLDisabled() || skipSSL
	}

	if endpoint == "" {
		endpoint = cmd.ui.Ask(T("API endpoint"))
	} else {
		cmd.ui.Say(T("API endpoint: {{.Endpoint}}", map[string]interface{}{"Endpoint": terminal.EntityNameColor(endpoint)}))
	}

	return endpoint, skipSSL
}

func (cmd Login) authenticateSSO(c flags.FlagContext) {
	prompts, err := cmd.authenticator.GetLoginPromptsAndSaveUAAServerURL()
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	credentials := make(map[string]string)
	passcode := prompts["passcode"]

	for i := 0; i < maxLoginTries; i++ {
		credentials["passcode"] = cmd.ui.AskForPassword("%s", passcode.DisplayName)

		cmd.ui.Say(T("Authenticating..."))
		err = cmd.authenticator.Authenticate(credentials)

		if err == nil {
			cmd.ui.Ok()
			cmd.ui.Say("")
			break
		}

		cmd.ui.Say(err.Error())
	}

	if err != nil {
		cmd.ui.Failed(T("Unable to authenticate."))
	}
}

func (cmd Login) authenticate(c flags.FlagContext) {
	usernameFlagValue := c.String("u")
	passwordFlagValue := c.String("p")

	prompts, err := cmd.authenticator.GetLoginPromptsAndSaveUAAServerURL()
	if err != nil {
		cmd.ui.Failed(err.Error())
	}
	passwordKeys := []string{}
	credentials := make(map[string]string)

	if value, ok := prompts["username"]; ok {
		if prompts["username"].Type == core_config.AuthPromptTypeText && usernameFlagValue != "" {
			credentials["username"] = usernameFlagValue
		} else {
			credentials["username"] = cmd.ui.Ask("%s", value.DisplayName)
		}
	}

	for key, prompt := range prompts {
		if prompt.Type == core_config.AuthPromptTypePassword {
			if key == "passcode" {
				continue
			}

			passwordKeys = append(passwordKeys, key)
		} else if key == "username" {
			continue
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

		cmd.ui.Say(T("Authenticating..."))
		err = cmd.authenticator.Authenticate(credentials)

		if err == nil {
			cmd.ui.Ok()
			cmd.ui.Say("")
			break
		}

		cmd.ui.Say(err.Error())
	}

	if err != nil {
		cmd.ui.Failed(T("Unable to authenticate."))
	}
}

func (cmd Login) setOrganization(c flags.FlagContext) (isOrgSet bool) {
	orgName := c.String("o")

	if orgName == "" {
		availableOrgs := []models.Organization{}
		orgs, apiErr := cmd.orgRepo.ListOrgs()
		if apiErr != nil {
			cmd.ui.Failed(T("Error finding available orgs\n{{.ApiErr}}",
				map[string]interface{}{"ApiErr": apiErr.Error()}))
		}
		for _, org := range orgs {
			if len(availableOrgs) < maxChoices {
				availableOrgs = append(availableOrgs, org)
			}
		}

		if len(availableOrgs) == 0 {
			return false
		} else if len(availableOrgs) == 1 {
			cmd.targetOrganization(availableOrgs[0])
			return true
		} else {
			orgName = cmd.promptForOrgName(availableOrgs)
			if orgName == "" {
				cmd.ui.Say("")
				return false
			}
		}
	}

	org, err := cmd.orgRepo.FindByName(orgName)
	if err != nil {
		cmd.ui.Failed(T("Error finding org {{.OrgName}}\n{{.Err}}",
			map[string]interface{}{"OrgName": terminal.EntityNameColor(orgName), "Err": err.Error()}))
	}

	cmd.targetOrganization(org)
	return true
}

func (cmd Login) promptForOrgName(orgs []models.Organization) string {
	orgNames := []string{}
	for _, org := range orgs {
		orgNames = append(orgNames, org.Name)
	}

	return cmd.promptForName(orgNames, T("Select an org (or press enter to skip):"), "Org")
}

func (cmd Login) targetOrganization(org models.Organization) {
	cmd.config.SetOrganizationFields(org.OrganizationFields)
	cmd.ui.Say(T("Targeted org {{.OrgName}}\n",
		map[string]interface{}{"OrgName": terminal.EntityNameColor(org.Name)}))
}

func (cmd Login) setSpace(c flags.FlagContext) {
	spaceName := c.String("s")

	if spaceName == "" {
		var availableSpaces []models.Space
		err := cmd.spaceRepo.ListSpaces(func(space models.Space) bool {
			availableSpaces = append(availableSpaces, space)
			return (len(availableSpaces) < maxChoices)
		})
		if err != nil {
			cmd.ui.Failed(T("Error finding available spaces\n{{.Err}}",
				map[string]interface{}{"Err": err.Error()}))
		}

		if len(availableSpaces) == 0 {
			return
		} else if len(availableSpaces) == 1 {
			cmd.targetSpace(availableSpaces[0])
			return
		} else {
			spaceName = cmd.promptForSpaceName(availableSpaces)
			if spaceName == "" {
				cmd.ui.Say("")
				return
			}
		}
	}

	space, err := cmd.spaceRepo.FindByName(spaceName)
	if err != nil {
		cmd.ui.Failed(T("Error finding space {{.SpaceName}}\n{{.Err}}",
			map[string]interface{}{"SpaceName": terminal.EntityNameColor(spaceName), "Err": err.Error()}))
	}

	cmd.targetSpace(space)
}

func (cmd Login) promptForSpaceName(spaces []models.Space) string {
	spaceNames := []string{}
	for _, space := range spaces {
		spaceNames = append(spaceNames, space.Name)
	}

	return cmd.promptForName(spaceNames, T("Select a space (or press enter to skip):"), "Space")
}

func (cmd Login) targetSpace(space models.Space) {
	cmd.config.SetSpaceFields(space.SpaceFields)
	cmd.ui.Say(T("Targeted space {{.SpaceName}}\n",
		map[string]interface{}{"SpaceName": terminal.EntityNameColor(space.Name)}))
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
			cmd.ui.Say(T("There are too many options to display, please type in the name."))
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
