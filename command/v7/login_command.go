package v7

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"

	"code.cloudfoundry.org/cli/api/uaa"
	"code.cloudfoundry.org/cli/util/ui"
	"code.cloudfoundry.org/clock"

	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/uaa/constant"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v7/shared"
)

//go:generate counterfeiter . ActorReloader

type ActorReloader interface {
	Reload(command.Config, command.UI) (Actor, error)
}

type ActualActorReloader struct{}

func (a ActualActorReloader) Reload(config command.Config, ui command.UI) (Actor, error) {
	ccClient, uaaClient, err := shared.GetNewClientsAndConnectToCF(config, ui, "")
	if err != nil {
		return nil, err
	}

	return v7action.NewActor(ccClient, config, nil, uaaClient, clock.NewClock()), nil
}

const maxLoginTries = 3

type LoginCommand struct {
	UI            command.UI
	Actor         Actor
	Config        command.Config
	ActorReloader ActorReloader

	APIEndpoint       string      `short:"a" description:"API endpoint (e.g. https://api.example.com)"`
	Organization      string      `short:"o" description:"Org"`
	Password          string      `short:"p" description:"Password"`
	Space             string      `short:"s" description:"Space"`
	SkipSSLValidation bool        `long:"skip-ssl-validation" description:"Skip verification of the API endpoint. Not recommended!"`
	SSO               bool        `long:"sso" description:"Prompt for a one-time passcode to login"`
	SSOPasscode       string      `long:"sso-passcode" description:"One-time passcode"`
	Username          string      `short:"u" description:"Username"`
	Origin            string      `long:"origin" description:"Indicates the identity provider to be used for login"`
	usage             interface{} `usage:"CF_NAME login [-a API_URL] [-u USERNAME] [-p PASSWORD] [-o ORG] [-s SPACE] [--sso | --sso-passcode PASSCODE] [--origin ORIGIN]\n\nWARNING:\n   Providing your password as a command line option is highly discouraged\n   Your password may be visible to others and may be recorded in your shell history\n\nEXAMPLES:\n   CF_NAME login (omit username and password to login interactively -- CF_NAME will prompt for both)\n   CF_NAME login -u name@example.com -p pa55woRD (specify username and password as arguments)\n   CF_NAME login -u name@example.com -p \"my password\" (use quotes for passwords with a space)\n   CF_NAME login -u name@example.com -p \"\\\"password\\\"\" (escape quotes if used in password)\n   CF_NAME login --sso (CF_NAME will provide a url to obtain a one-time passcode to login)\n   CF_NAME login --origin ldap"`
	relatedCommands   interface{} `related_commands:"api, auth, target"`
}

func (cmd *LoginCommand) Setup(config command.Config, ui command.UI) error {
	ccClient, _ := shared.NewWrappedCloudControllerClient(config, ui)
	cmd.Actor = v7action.NewActor(ccClient, config, nil, nil, clock.NewClock())
	cmd.ActorReloader = ActualActorReloader{}

	cmd.UI = ui
	cmd.Config = config
	return nil
}

func (cmd *LoginCommand) Execute(args []string) error {
	if cmd.Config.UAAGrantType() == string(constant.GrantTypeClientCredentials) {
		return translatableerror.PasswordGrantTypeLogoutRequiredError{}
	}

	if cmd.Config.UAAOAuthClient() != "cf" || cmd.Config.UAAOAuthClientSecret() != "" {
		return translatableerror.ManualClientCredentialsError{}
	}

	err := cmd.validateFlags()
	if err != nil {
		return err
	}

	endpoint, err := cmd.determineAPIEndpoint()
	if err != nil {
		return err
	}

	err = cmd.targetAPI(endpoint)
	if err != nil {
		translatedErr := translatableerror.ConvertToTranslatableError(err)
		if invalidSSLErr, ok := translatedErr.(translatableerror.InvalidSSLCertError); ok {
			invalidSSLErr.SuggestedCommand = "login"
			return invalidSSLErr
		}
		return err
	}

	versionWarning, err := shared.CheckCCAPIVersion(cmd.Config.APIVersion())
	if err != nil {
		cmd.UI.DisplayWarning("Warning: unable to determine whether targeted API's version meets minimum supported.")
	}
	if versionWarning != "" {
		cmd.UI.DisplayWarning(versionWarning)
	}

	cmd.UI.DisplayNewline()

	cmd.Actor, err = cmd.ActorReloader.Reload(cmd.Config, cmd.UI)
	if err != nil {
		return err
	}

	defer cmd.showStatus()

	var authErr error
	if cmd.SSO || cmd.SSOPasscode != "" {
		authErr = cmd.authenticateSSO()
	} else {
		authErr = cmd.authenticate()
	}

	if authErr != nil {
		return errors.New("Unable to authenticate.")
	}

	err = cmd.Config.WriteConfig()
	if err != nil {
		return fmt.Errorf("Error writing config: %s", err.Error())
	}

	if cmd.Organization != "" {
		org, warnings, err := cmd.Actor.GetOrganizationByName(cmd.Organization)
		cmd.UI.DisplayWarnings(warnings)
		if err != nil {
			return err
		}

		cmd.Config.SetOrganizationInformation(org.GUID, org.Name)
	} else {
		orgs, warnings, err := cmd.Actor.GetOrganizations("")
		cmd.UI.DisplayWarnings(warnings)
		if err != nil {
			return err
		}

		if len(orgs) == 1 {
			cmd.Config.SetOrganizationInformation(orgs[0].GUID, orgs[0].Name)
		} else if len(orgs) > 1 {
			chosenOrg, err := cmd.promptChosenOrg(orgs)
			if err != nil {
				return err
			}

			if chosenOrg.GUID != "" {
				cmd.Config.SetOrganizationInformation(chosenOrg.GUID, chosenOrg.Name)
			}
		}
	}

	targetedOrg := cmd.Config.TargetedOrganization()

	if targetedOrg.GUID != "" {
		cmd.UI.DisplayTextWithFlavor("Targeted org {{.Organization}}.", map[string]interface{}{
			"Organization": cmd.Config.TargetedOrganizationName(),
		})
		cmd.UI.DisplayNewline()

		if cmd.Space != "" {
			space, warnings, err := cmd.Actor.GetSpaceByNameAndOrganization(cmd.Space, targetedOrg.GUID)
			cmd.UI.DisplayWarnings(warnings)
			if err != nil {
				return err
			}
			cmd.targetSpace(space)
		} else {
			spaces, warnings, err := cmd.Actor.GetOrganizationSpaces(targetedOrg.GUID)
			cmd.UI.DisplayWarnings(warnings)
			if err != nil {
				return err
			}

			if len(spaces) == 1 {
				cmd.targetSpace(spaces[0])
			} else if len(spaces) > 1 {
				chosenSpace, err := cmd.promptChosenSpace(spaces)
				if err != nil {
					return err
				}
				if chosenSpace.Name != "" {
					cmd.targetSpace(chosenSpace)
				}
			}
		}
	}

	return nil
}

func (cmd *LoginCommand) determineAPIEndpoint() (v7action.TargetSettings, error) {
	endpoint := cmd.APIEndpoint
	skipSSLValidation := cmd.SkipSSLValidation

	var configTarget = cmd.Config.Target()

	if endpoint == "" && configTarget != "" {
		endpoint = configTarget
		skipSSLValidation = cmd.Config.SkipSSLValidation() || cmd.SkipSSLValidation
	}

	if len(endpoint) > 0 {
		cmd.UI.DisplayTextWithFlavor("API endpoint: {{.APIEndpoint}}", map[string]interface{}{
			"APIEndpoint": endpoint,
		})
	} else {
		userInput, err := cmd.UI.DisplayTextPrompt("API endpoint")
		if err != nil {
			return v7action.TargetSettings{}, err
		}
		endpoint = userInput
	}

	strippedEndpoint := strings.TrimRight(endpoint, "/")
	parsedURL, err := url.Parse(strippedEndpoint)
	if err != nil {
		return v7action.TargetSettings{}, err
	}
	if parsedURL.Scheme == "" {
		parsedURL.Scheme = "https"
	}

	return v7action.TargetSettings{URL: parsedURL.String(), SkipSSLValidation: skipSSLValidation}, nil
}

func (cmd *LoginCommand) targetAPI(settings v7action.TargetSettings) error {
	warnings, err := cmd.Actor.SetTarget(settings)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	if strings.HasPrefix(settings.URL, "http:") {
		cmd.UI.DisplayWarning("Warning: Insecure http API endpoint detected: secure https API endpoints are recommended")
	}

	return nil
}

func (cmd *LoginCommand) authenticate() error {
	var err error
	var credentials = make(map[string]string)

	prompts := cmd.Actor.GetLoginPrompts()
	nonPasswordPrompts, passwordPrompts := cmd.groupPrompts(prompts)

	if value, ok := prompts["username"]; ok {
		credentials["username"], err = cmd.getFlagValOrPrompt(&cmd.Username, value, true)
		if err != nil {
			return err
		}
	}

	for key, prompt := range nonPasswordPrompts {
		credentials[key], err = cmd.UI.DisplayTextPrompt(prompt.DisplayName)
		if err != nil {
			return err
		}
	}

	for i := 0; i < maxLoginTries; i++ {
		// ensure that password gets prompted before other codes (eg. mfa code)
		if prompt, ok := prompts["password"]; ok {
			credentials["password"], err = cmd.getFlagValOrPrompt(&cmd.Password, prompt, false)
			if err != nil {
				return err
			}
		}

		for key, prompt := range passwordPrompts {
			credentials[key], err = cmd.UI.DisplayPasswordPrompt(prompt.DisplayName)
			if err != nil {
				return err
			}
		}

		cmd.UI.DisplayNewline()
		cmd.UI.DisplayText("Authenticating...")

		err = cmd.Actor.Authenticate(credentials, cmd.Origin, constant.GrantTypePassword)

		if err != nil {
			cmd.UI.DisplayWarning(translatableerror.ConvertToTranslatableError(err).Error())
			cmd.UI.DisplayNewline()

			if _, ok := err.(uaa.AccountLockedError); ok {
				break
			}
		}

		if err == nil {
			cmd.UI.DisplayOK()
			break
		}
	}

	return err
}

func (cmd *LoginCommand) authenticateSSO() error {
	prompts := cmd.Actor.GetLoginPrompts()

	var err error
	for i := 0; i < maxLoginTries; i++ {
		var passcode string

		passcode, err = cmd.getFlagValOrPrompt(&cmd.SSOPasscode, prompts["passcode"], false)
		if err != nil {
			return err
		}

		credentials := map[string]string{"passcode": passcode}

		cmd.UI.DisplayText("Authenticating...")
		err = cmd.Actor.Authenticate(credentials, "", constant.GrantTypePassword)

		if err != nil {
			cmd.UI.DisplayWarning(translatableerror.ConvertToTranslatableError(err).Error())
			cmd.UI.DisplayNewline()
		} else {
			cmd.UI.DisplayOK()
			cmd.UI.DisplayNewline()
			break
		}
	}
	return err
}

func (cmd *LoginCommand) groupPrompts(prompts map[string]coreconfig.AuthPrompt) (map[string]coreconfig.AuthPrompt, map[string]coreconfig.AuthPrompt) {
	var (
		nonPasswordPrompts = make(map[string]coreconfig.AuthPrompt)
		passwordPrompts    = make(map[string]coreconfig.AuthPrompt)
	)

	for key, prompt := range prompts {
		if prompt.Type == coreconfig.AuthPromptTypePassword {
			if key == "passcode" || key == "password" {
				continue
			}

			passwordPrompts[key] = prompt
		} else {
			if key == "username" {
				continue
			}

			nonPasswordPrompts[key] = prompt
		}
	}

	return nonPasswordPrompts, passwordPrompts
}

func (cmd *LoginCommand) getFlagValOrPrompt(field *string, prompt coreconfig.AuthPrompt, isText bool) (string, error) {
	if *field != "" {
		value := *field
		*field = ""
		return value, nil
	} else {
		if isText {
			return cmd.UI.DisplayTextPrompt(prompt.DisplayName)
		}
		return cmd.UI.DisplayPasswordPrompt(prompt.DisplayName)
	}
}

func (cmd *LoginCommand) showStatus() {
	tableContent := [][]string{
		{
			cmd.UI.TranslateText("API endpoint:"),
			strings.TrimRight(cmd.Config.Target(), "/"),
		},
		{
			cmd.UI.TranslateText("API version:"),
			cmd.Config.APIVersion(),
		},
	}

	user, err := cmd.Config.CurrentUserName()
	if user == "" || err != nil {
		cmd.UI.DisplayKeyValueTable("", tableContent, 3)
		command.DisplayNotLoggedInText(cmd.Config.BinaryName(), cmd.UI)
		return
	}
	tableContent = append(tableContent, []string{cmd.UI.TranslateText("user:"), user})

	orgName := cmd.Config.TargetedOrganizationName()
	if orgName == "" {
		cmd.UI.DisplayKeyValueTable("", tableContent, 3)
		cmd.UI.DisplayText("No org or space targeted, use '{{.CFTargetCommand}} -o ORG -s SPACE'",
			map[string]interface{}{
				"CFTargetCommand": fmt.Sprintf("%s target", cmd.Config.BinaryName()),
			},
		)
		return
	}
	tableContent = append(tableContent, []string{cmd.UI.TranslateText("org:"), orgName})

	spaceContent := cmd.Config.TargetedSpace().Name
	if spaceContent == "" {
		spaceContent = cmd.UI.TranslateText("No space targeted, use '{{.Command}}'",
			map[string]interface{}{
				"Command": fmt.Sprintf("%s target -s SPACE", cmd.Config.BinaryName()),
			},
		)
	}
	tableContent = append(tableContent, []string{cmd.UI.TranslateText("space:"), spaceContent})

	cmd.UI.DisplayKeyValueTable("", tableContent, 3)
}

func (cmd *LoginCommand) promptChosenOrg(orgs []v7action.Organization) (v7action.Organization, error) {
	orgNames := make([]string, len(orgs))
	for i, org := range orgs {
		orgNames[i] = org.Name
	}

	chosenOrgName, err := cmd.promptMenu(orgNames, "Select an org:", "Org")

	if err != nil {
		if invalidChoice, ok := err.(ui.InvalidChoiceError); ok {
			return v7action.Organization{}, translatableerror.OrganizationNotFoundError{Name: invalidChoice.Choice}
		}

		if err == io.EOF {
			return v7action.Organization{}, nil
		}

		return v7action.Organization{}, err
	}

	for _, org := range orgs {
		if org.Name == chosenOrgName {
			return org, nil
		}
	}

	return v7action.Organization{}, nil
}

func (cmd *LoginCommand) promptChosenSpace(spaces []v7action.Space) (v7action.Space, error) {
	spaceNames := make([]string, len(spaces))
	for i, space := range spaces {
		spaceNames[i] = space.Name
	}

	chosenSpaceName, err := cmd.promptMenu(spaceNames, "Select a space:", "Space")
	if err != nil {
		if invalidChoice, ok := err.(ui.InvalidChoiceError); ok {
			return v7action.Space{}, translatableerror.SpaceNotFoundError{Name: invalidChoice.Choice}
		}

		if err == io.EOF {
			return v7action.Space{}, nil
		}

		return v7action.Space{}, err
	}

	for _, space := range spaces {
		if space.Name == chosenSpaceName {
			return space, nil
		}
	}
	return v7action.Space{}, nil
}

func (cmd *LoginCommand) promptMenu(choices []string, text string, prompt string) (string, error) {
	var choice string
	var err error

	if len(choices) < 50 {
		for {
			cmd.UI.DisplayText(text)
			choice, err = cmd.UI.DisplayTextMenu(choices, prompt)
			if err != ui.ErrInvalidIndex {
				break
			}
		}
	} else {
		cmd.UI.DisplayText(text)
		cmd.UI.DisplayText("There are too many options to display; please type in the name.")
		cmd.UI.DisplayNewline()
		defaultChoice := "enter to skip"
		choice, err = cmd.UI.DisplayOptionalTextPrompt(defaultChoice, prompt)

		if choice == defaultChoice {
			return "", nil
		}
		if !contains(choices, choice) {
			return "", ui.InvalidChoiceError{Choice: choice}
		}
	}

	return choice, err
}

func (cmd *LoginCommand) targetSpace(space v7action.Space) {
	cmd.Config.SetSpaceInformation(space.GUID, space.Name, true)

	cmd.UI.DisplayTextWithFlavor("Targeted space {{.Space}}.", map[string]interface{}{
		"Space": space.Name,
	})
	cmd.UI.DisplayNewline()
}

func (cmd *LoginCommand) validateFlags() error {
	if cmd.Origin != "" && cmd.SSO {
		return translatableerror.ArgumentCombinationError{
			Args: []string{"--sso", "--origin"},
		}
	}

	if cmd.Origin != "" && cmd.SSOPasscode != "" {
		return translatableerror.ArgumentCombinationError{
			Args: []string{"--sso-passcode", "--origin"},
		}
	}

	if cmd.SSO && cmd.SSOPasscode != "" {
		return translatableerror.ArgumentCombinationError{
			Args: []string{"--sso-passcode", "--sso"},
		}
	}

	return nil
}

func contains(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}
