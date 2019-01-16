package v6

import (
	"net/url"

	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v6/shared"
)

//go:generate counterfeiter . LoginActor

type LoginActor interface {
	SetTarget(settings v2action.TargetSettings) (v2action.Warnings, error)
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

	UI     command.UI
	Actor  LoginActor
	Config command.Config
}

func (cmd *LoginCommand) Setup(config command.Config, ui command.UI) error {
	ccClient, uaaClient, err := shared.NewClients(config, ui, false)
	if err != nil {
		return err
	}
	cmd.Actor = v2action.NewActor(ccClient, uaaClient, config)
	cmd.UI = ui
	cmd.Config = config
	return nil
}

func (cmd *LoginCommand) Execute(args []string) error {
	if cmd.Config.Experimental() {
		var err error
		if cmd.APIEndpoint == "" {
			apiEndpoint, err := cmd.UI.DisplayTextPrompt("API endpoint")
			if err != nil {
				return err
			}
			cmd.APIEndpoint = apiEndpoint
		} else {
			cmd.UI.DisplayText("API endpoint: {{.APIEndpoint}}", map[string]interface{}{
				"APIEndpoint": cmd.APIEndpoint,
			})
		}
		endpoint, _ := url.Parse(cmd.APIEndpoint)
		if endpoint.Scheme == "" {
			endpoint.Scheme = "https"
		}
		settings := v2action.TargetSettings{
			URL:               endpoint.String(),
			SkipSSLValidation: true,
		}
		_, err = cmd.Actor.SetTarget(settings)
		if err != nil {
			return err
		}

		cmd.UI.DisplayText("API endpoint: {{.APIEndpoint}} (API Version: {{.APIVersion}})", map[string]interface{}{
			"APIEndpoint": cmd.APIEndpoint,
			"APIVersion":  cmd.Config.APIVersion(),
		})
		return nil
	}
	return translatableerror.UnrefactoredCommandError{}
}
