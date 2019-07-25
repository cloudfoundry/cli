package v6

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"

	"code.cloudfoundry.org/cli/api/uaa"
	"code.cloudfoundry.org/cli/util/ui"

	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/api/uaa/constant"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v6/shared"
)

//go:generate counterfeiter . LoginActor

const maxLoginTries = 3

type LoginActor interface {
	Authenticate(credentials map[string]string, origin string, grantType constant.GrantType) error
	GetLoginPrompts() map[string]coreconfig.AuthPrompt
	GetOrganizationByName(orgName string) (v3action.Organization, v3action.Warnings, error)
	GetOrganizationSpaces(orgName string) ([]v3action.Space, v3action.Warnings, error)
	GetSpaceByNameAndOrganization(spaceName string, orgGUID string) (v3action.Space, v3action.Warnings, error)
	GetOrganizations() ([]v3action.Organization, v3action.Warnings, error)
	SetTarget(settings v3action.TargetSettings) (v3action.Warnings, error)
}

//go:generate counterfeiter . VersionChecker

type VersionChecker interface {
	MinCLIVersion() string
	CloudControllerAPIVersion() string
}

//go:generate counterfeiter . ActorMaker

type ActorMaker interface {
	NewActor(command.Config, command.UI, bool) (LoginActor, error)
}

//go:generate counterfeiter . CheckerMaker

type CheckerMaker interface {
	NewVersionChecker(command.Config, command.UI, bool) (VersionChecker, error)
}

type ActorMakerFunc func(command.Config, command.UI, bool) (LoginActor, error)
type CheckerMakerFunc func(command.Config, command.UI, bool) (VersionChecker, error)

func (a ActorMakerFunc) NewActor(config command.Config, ui command.UI, targetCF bool) (LoginActor, error) {
	return a(config, ui, targetCF)
}

func (c CheckerMakerFunc) NewVersionChecker(config command.Config, ui command.UI, targetCF bool) (VersionChecker, error) {
	return c(config, ui, targetCF)
}

var actorMaker ActorMakerFunc = func(config command.Config, ui command.UI, targetCF bool) (LoginActor, error) {
	client, uaa, err := shared.NewV3BasedClients(config, ui, targetCF, "")
	if err != nil {
		return nil, err
	}

	v3Actor := v3action.NewActor(client, config, nil, uaa)
	return v3Actor, nil
}

var checkerMaker CheckerMakerFunc = func(config command.Config, ui command.UI, targetCF bool) (VersionChecker, error) {
	client, uaa, err := shared.NewClients(config, ui, targetCF)
	if err != nil {
		return nil, err
	}

	v2Actor := v2action.NewActor(client, uaa, config)
	return v2Actor, nil
}

type LoginCommand struct {
	APIEndpoint       string      `short:"a" description:"API endpoint (e.g. https://api.example.com)"`
	Organization      string      `short:"o" description:"Org"`
	Password          string      `short:"p" description:"Password"`
	Space             string      `short:"s" description:"Space"`
	SkipSSLValidation bool        `long:"skip-ssl-validation" description:"Skip verification of the API endpoint. Not recommended!"`
	SSO               bool        `long:"sso" description:"Prompt for a one-time passcode to login"`
	SSOPasscode       string      `long:"sso-passcode" description:"One-time passcode"`
	Username          string      `short:"u" description:"Username"`
	usage             interface{} `usage:"CF_NAME login [-a API_URL] [-u USERNAME] [-p PASSWORD] [-o ORG] [-s SPACE] [--sso | --sso-passcode PASSCODE]\n\nWARNING:\n   Providing your password as a command line option is highly discouraged\n   Your password may be visible to others and may be recorded in your shell history\n\nEXAMPLES:\n   CF_NAME login (omit username and password to login interactively -- CF_NAME will prompt for both)\n   CF_NAME login -u name@example.com -p pa55woRD (specify username and password as arguments)\n   CF_NAME login -u name@example.com -p \"my password\" (use quotes for passwords with a space)\n   CF_NAME login -u name@example.com -p \"\\\"password\\\"\" (escape quotes if used in password)\n   CF_NAME login --sso (CF_NAME will provide a url to obtain a one-time passcode to login)"`
	relatedCommands   interface{} `related_commands:"api, auth, target"`

	UI           command.UI
	Actor        LoginActor
	ActorMaker   ActorMaker
	Checker      VersionChecker
	CheckerMaker CheckerMaker
	Config       command.Config
}

func (cmd *LoginCommand) Setup(config command.Config, ui command.UI) error {
	cmd.ActorMaker = actorMaker
	actor, err := cmd.ActorMaker.NewActor(config, ui, false)
	if err != nil {
		return err
	}
	cmd.CheckerMaker = checkerMaker
	cmd.Actor = actor
	cmd.UI = ui
	cmd.Config = config
	return nil
}

func (cmd *LoginCommand) Execute(args []string) error {
	if !cmd.Config.ExperimentalLogin() {
		return translatableerror.UnrefactoredCommandError{}
	}
	cmd.UI.DisplayWarning("Using experimental login command, some behavior may be different")

	err := cmd.getAPI()
	if err != nil {
		return err
	}

	cmd.UI.DisplayNewline()

	err = cmd.retargetAPI()
	if err != nil {
		return err
	}

	defer cmd.showStatus()

	if cmd.Config.UAAGrantType() == string(constant.GrantTypeClientCredentials) {
		return errors.New("Service account currently logged in. Use 'cf logout' to log out service account and try again.")
	} else if cmd.Config.UAAOAuthClient() != "cf" || cmd.Config.UAAOAuthClientSecret() != "" {
		cmd.UI.DisplayWarning("Deprecation warning: Manually writing your client credentials to the config.json is deprecated and will be removed in the future. For similar functionality, please use the `cf auth --client-credentials` command instead.")
	}

	var authErr error
	if cmd.SSO || cmd.SSOPasscode != "" {
		if cmd.SSO && cmd.SSOPasscode != "" {
			return translatableerror.ArgumentCombinationError{Args: []string{"--sso-passcode", "--sso"}}
		}
		authErr = cmd.authenticateSSO()
	} else {
		authErr = cmd.authenticate()
	}

	if authErr != nil {
		return errors.New("Unable to authenticate.")
	}

	if cmd.Organization != "" {
		org, warnings, err := cmd.Actor.GetOrganizationByName(cmd.Organization)
		cmd.UI.DisplayWarnings(warnings)
		if err != nil {
			return err
		}
		cmd.Config.SetOrganizationInformation(org.GUID, org.Name)
	} else {
		orgs, warnings, err := cmd.Actor.GetOrganizations()
		cmd.UI.DisplayWarnings(warnings)
		if err != nil {
			return err
		}
		if len(orgs) == 1 {
			cmd.Config.SetOrganizationInformation(orgs[0].GUID, orgs[0].Name)
		} else if len(orgs) > 1 {
			var emptyOrg v3action.Organization
			chosenOrg, err := cmd.promptChosenOrg(orgs)
			if err != nil {
				return err
			}
			if chosenOrg != emptyOrg {
				cmd.Config.SetOrganizationInformation(chosenOrg.GUID, chosenOrg.Name)
			}
		}
	}

	targetedOrg := cmd.Config.TargetedOrganization()

	if targetedOrg.GUID != "" {
		cmd.UI.DisplayTextWithFlavor("Targeted org {{.Organization}}", map[string]interface{}{
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
				var emptySpace v3action.Space
				chosenSpace, err := cmd.promptChosenSpace(spaces)
				if err != nil {
					return err
				}
				if chosenSpace != emptySpace {
					cmd.targetSpace(chosenSpace)
				}
			}
		}
	}

	err = cmd.checkMinCLIVersion()
	if err != nil {
		return err
	}

	cmd.UI.DisplayNewline()
	cmd.UI.DisplayNewline()
	cmd.UI.DisplayNewline()

	return nil
}

func (cmd *LoginCommand) targetSpace(space v3action.Space) {
	cmd.Config.SetSpaceInformation(space.GUID, space.Name, true)

	cmd.UI.DisplayTextWithFlavor("Targeted space {{.Space}}", map[string]interface{}{
		"Space": space.Name,
	})
}

func (cmd *LoginCommand) getAPI() error {
	if cmd.APIEndpoint != "" {
		cmd.UI.DisplayTextWithFlavor("API endpoint: {{.APIEndpoint}}", map[string]interface{}{
			"APIEndpoint": cmd.APIEndpoint,
		})
	} else if cmd.Config.Target() != "" {
		cmd.APIEndpoint = cmd.Config.Target()
		cmd.UI.DisplayTextWithFlavor("API endpoint: {{.APIEndpoint}}", map[string]interface{}{
			"APIEndpoint": cmd.APIEndpoint,
		})
	} else {
		apiEndpoint, err := cmd.UI.DisplayTextPrompt("API endpoint")
		if err != nil {
			return err
		}
		cmd.APIEndpoint = apiEndpoint
	}
	return nil
}

func (cmd *LoginCommand) retargetAPI() error {
	strippedEndpoint := strings.TrimRight(cmd.APIEndpoint, "/")
	endpoint, _ := url.Parse(strippedEndpoint)
	if endpoint.Scheme == "" {
		endpoint.Scheme = "https"
	}

	settings := v3action.TargetSettings{
		URL:               endpoint.String(),
		SkipSSLValidation: cmd.Config.SkipSSLValidation() || cmd.SkipSSLValidation,
	}
	_, err := cmd.Actor.SetTarget(settings)
	if err != nil {
		return err
	}

	return cmd.reloadActor()
}

func (cmd *LoginCommand) authenticate() error {
	prompts := cmd.Actor.GetLoginPrompts()
	credentials := make(map[string]string)

	if value, ok := prompts["username"]; ok {
		if prompts["username"].Type == coreconfig.AuthPromptTypeText && cmd.Username != "" {
			credentials["username"] = cmd.Username
		} else {
			var err error
			credentials["username"], err = cmd.UI.DisplayTextPrompt(value.DisplayName)
			if err != nil {
				return err
			}
			cmd.UI.DisplayNewline()
		}
	}

	passwordKeys := []string{}
	for key, prompt := range prompts {
		if prompt.Type == coreconfig.AuthPromptTypePassword {
			if key == "passcode" || key == "password" {
				continue
			}

			passwordKeys = append(passwordKeys, key)
		} else if key == "username" {
			continue
		} else {
			var err error
			credentials[key], err = cmd.UI.DisplayTextPrompt(prompt.DisplayName)
			if err != nil {
				return err
			}
			cmd.UI.DisplayNewline()
		}
	}

	var err error
	for i := 0; i < maxLoginTries; i++ {
		var promptedCredentials map[string]string
		promptedCredentials, err = cmd.passwordPrompts(prompts, credentials, passwordKeys)
		if err != nil {
			return err
		}

		cmd.UI.DisplayText("Authenticating...")

		err = cmd.Actor.Authenticate(promptedCredentials, "", constant.GrantTypePassword)

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
	if err != nil {
		return err
	}
	return nil
}

func (cmd *LoginCommand) authenticateSSO() error {
	prompts := cmd.Actor.GetLoginPrompts()
	credentials := make(map[string]string)

	var err error
	for i := 0; i < maxLoginTries; i++ {
		if len(cmd.SSOPasscode) > 0 {
			credentials["passcode"] = cmd.SSOPasscode
			cmd.SSOPasscode = ""
		} else {
			credentials["passcode"], err = cmd.UI.DisplayPasswordPrompt(prompts["passcode"].DisplayName)
			if err != nil {
				return err
			}
		}

		credentialsCopy := make(map[string]string, len(credentials))
		for k, v := range credentials {
			credentialsCopy[k] = v
		}

		cmd.UI.DisplayText("Authenticating...")
		err = cmd.Actor.Authenticate(credentialsCopy, "", constant.GrantTypePassword)

		if err != nil {
			cmd.UI.DisplayWarning(translatableerror.ConvertToTranslatableError(err).Error())
			cmd.UI.DisplayNewline()
		}

		if err == nil {
			cmd.UI.DisplayOK()
			cmd.UI.DisplayNewline()
			break
		}
	}
	if err != nil {
		return err
	}
	return nil
}

func (cmd *LoginCommand) checkMinCLIVersion() error {
	newChecker, err := cmd.CheckerMaker.NewVersionChecker(cmd.Config, cmd.UI, true)
	if err != nil {
		return err
	}

	cmd.Checker = newChecker
	cmd.Config.SetMinCLIVersion(cmd.Checker.MinCLIVersion())
	return command.WarnIfCLIVersionBelowAPIDefinedMinimum(cmd.Config, cmd.Checker.CloudControllerAPIVersion(), cmd.UI)
}

func (cmd *LoginCommand) passwordPrompts(prompts map[string]coreconfig.AuthPrompt, credentials map[string]string, passwordKeys []string) (map[string]string, error) {
	// ensure that password gets prompted before other codes (eg. mfa code)
	var err error
	if passPrompt, ok := prompts["password"]; ok {
		if cmd.Password != "" {
			credentials["password"] = cmd.Password
			cmd.Password = ""
		} else {
			credentials["password"], err = cmd.UI.DisplayPasswordPrompt(passPrompt.DisplayName)
			if err != nil {
				return nil, err
			}
		}
	}

	for _, key := range passwordKeys {
		cmd.UI.DisplayNewline()
		credentials[key], err = cmd.UI.DisplayPasswordPrompt(prompts[key].DisplayName)
		if err != nil {
			return nil, err
		}
	}

	credentialsCopy := make(map[string]string, len(credentials))
	for k, v := range credentials {
		credentialsCopy[k] = v
	}

	return credentialsCopy, nil
}

func (cmd *LoginCommand) reloadActor() error {
	newActor, err := cmd.ActorMaker.NewActor(cmd.Config, cmd.UI, true)
	if err != nil {
		return err
	}

	cmd.Actor = newActor

	return nil
}

func (cmd *LoginCommand) showStatus() {
	tableContent := [][]string{
		{
			cmd.UI.TranslateText("API endpoint:"),
			cmd.UI.TranslateText("{{.APIEndpoint}} (API version: {{.APIVersion}})",
				map[string]interface{}{
					"APIEndpoint": strings.TrimRight(cmd.APIEndpoint, "/"),
					"APIVersion":  cmd.Config.APIVersion(),
				}),
		},
	}

	user, err := cmd.Config.CurrentUserName()
	if user == "" || err != nil {
		cmd.UI.DisplayKeyValueTable("", tableContent, 3)
		cmd.displayNotLoggedIn()
		return
	}
	tableContent = append(tableContent, []string{cmd.UI.TranslateText("User:"), user})

	orgName := cmd.Config.TargetedOrganizationName()
	if orgName == "" {
		cmd.UI.DisplayKeyValueTable("", tableContent, 3)
		cmd.displayNotTargetted()
		return
	}
	tableContent = append(tableContent, []string{cmd.UI.TranslateText("Org:"), orgName})

	spaceName := cmd.Config.TargetedSpace().Name
	if spaceName == "" {
		tableContent = append(tableContent, []string{cmd.UI.TranslateText("Space:"),
			cmd.UI.TranslateText("No space targeted, use '{{.Command}}'", map[string]interface{}{
				"Command": fmt.Sprintf("%s target -s SPACE", cmd.Config.BinaryName()),
			})})
	} else {
		tableContent = append(tableContent, []string{cmd.UI.TranslateText("Space:"), spaceName})
	}

	cmd.UI.DisplayKeyValueTable("", tableContent, 3)
}

func (cmd *LoginCommand) displayNotLoggedIn() {
	cmd.UI.DisplayText(
		"Not logged in. Use '{{.CFLoginCommand}}' to log in.",
		map[string]interface{}{
			"CFLoginCommand": fmt.Sprintf("%s login", cmd.Config.BinaryName()),
		},
	)
}

func (cmd *LoginCommand) displayNotTargetted() {
	cmd.UI.DisplayText("No org or space targeted, use '{{.CFTargetCommand}} -o ORG -s SPACE'",
		map[string]interface{}{
			"CFTargetCommand": fmt.Sprintf("%s target", cmd.Config.BinaryName()),
		},
	)
}

func (cmd *LoginCommand) promptChosenOrg(orgs []v3action.Organization) (v3action.Organization, error) {
	orgNames := make([]string, len(orgs))
	for i, org := range orgs {
		orgNames[i] = org.Name
	}

	chosenOrgName, err := cmd.promptMenu(orgNames, "Select an org:", "Org")

	if err != nil {
		if invalidChoice, ok := err.(ui.InvalidChoiceError); ok {
			return v3action.Organization{}, translatableerror.OrganizationNotFoundError{Name: invalidChoice.Choice}
		} else if err == io.EOF {
			return v3action.Organization{}, nil
		}

		return v3action.Organization{}, err
	}

	if chosenOrgName == "" {
		return v3action.Organization{}, nil
	}

	for _, org := range orgs {
		if org.Name == chosenOrgName {
			return org, nil
		}
	}

	return v3action.Organization{}, translatableerror.OrganizationNotFoundError{Name: chosenOrgName}
}

func (cmd *LoginCommand) promptChosenSpace(spaces []v3action.Space) (v3action.Space, error) {
	spaceNames := make([]string, len(spaces))
	for i, space := range spaces {
		spaceNames[i] = space.Name
	}

	chosenSpaceName, err := cmd.promptMenu(spaceNames, "Select a space:", "Space")
	if err != nil {
		if invalidChoice, ok := err.(ui.InvalidChoiceError); ok {
			return v3action.Space{}, translatableerror.SpaceNotFoundError{Name: invalidChoice.Choice}
		} else if err == io.EOF {
			return v3action.Space{}, nil
		}
		return v3action.Space{}, nil
	}

	for _, space := range spaces {
		if space.Name == chosenSpaceName {
			return space, nil
		}
	}
	return v3action.Space{}, translatableerror.SpaceNotFoundError{Name: chosenSpaceName}
}

func (cmd *LoginCommand) promptMenu(choices []string, text string, prompt string) (string, error) {
	var (
		choice string
		err    error
	)

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
	}

	return choice, err
}
