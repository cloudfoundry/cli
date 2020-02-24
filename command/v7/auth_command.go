package v7

import (
	"fmt"

	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/uaa/constant"
	"code.cloudfoundry.org/cli/api/uaa/uaaversion"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/clock"
)

//go:generate counterfeiter . AuthActor

type AuthActor interface {
	Authenticate(credentials map[string]string, origin string, grantType constant.GrantType) error
	CloudControllerAPIVersion() string
	UAAAPIVersion() string
}

type AuthCommand struct {
	RequiredArgs      flag.Authentication `positional-args:"yes"`
	ClientCredentials bool                `long:"client-credentials" description:"Use (non-user) service account (also called client credentials)"`
	Origin            string              `long:"origin" description:"Indicates the identity provider to be used for authentication"`
	usage             interface{}         `usage:"CF_NAME auth USERNAME PASSWORD\n   CF_NAME auth USERNAME PASSWORD --origin ORIGIN\n   CF_NAME auth CLIENT_ID CLIENT_SECRET --client-credentials\n\nENVIRONMENT VARIABLES:\n   CF_USERNAME=user          Authenticating user. Overridden if USERNAME argument is provided.\n   CF_PASSWORD=password      Password associated with user. Overriden if PASSWORD argument is provided.\n\nWARNING:\n   Providing your password as a command line option is highly discouraged\n   Your password may be visible to others and may be recorded in your shell history\n   Consider using the CF_PASSWORD environment variable instead\n\nEXAMPLES:\n   CF_NAME auth name@example.com \"my password\" (use quotes for passwords with a space)\n   CF_NAME auth name@example.com \"\\\"password\\\"\" (escape quotes if used in password)"`
	relatedCommands   interface{}         `related_commands:"api, login, target"`

	UI     command.UI
	Config command.Config
	Actor  AuthActor
}

func (cmd *AuthCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config

	ccClient, uaaClient, err := shared.GetNewClientsAndConnectToCF(config, ui, "")
	if err != nil {
		return err
	}
	cmd.Actor = v7action.NewActor(ccClient, config, nil, uaaClient, clock.NewClock())

	return nil
}

func (cmd AuthCommand) Execute(args []string) error {
	if len(cmd.Origin) > 0 {
		err := command.MinimumUAAAPIVersionCheck(cmd.Actor.UAAAPIVersion(), uaaversion.MinUAAClientVersion, "Option '--origin'")
		if err != nil {
			return err
		}
	}

	if cmd.ClientCredentials && cmd.Origin != "" {
		return translatableerror.ArgumentCombinationError{
			Args: []string{"--client-credentials", "--origin"},
		}
	}

	username, password, err := cmd.getUsernamePassword()
	if err != nil {
		return err
	}

	if !cmd.ClientCredentials {
		if cmd.Config.UAAGrantType() == string(constant.GrantTypeClientCredentials) {
			return translatableerror.PasswordGrantTypeLogoutRequiredError{}
		} else if cmd.Config.UAAOAuthClient() != "cf" || cmd.Config.UAAOAuthClientSecret() != "" {
			cmd.UI.DisplayWarning("Deprecation warning: Manually writing your client credentials to the config.json is deprecated and will be removed in the future. For similar functionality, please use the `cf auth --client-credentials` command instead.")
		}
	}

	cmd.UI.DisplayTextWithFlavor(
		"API endpoint: {{.Endpoint}}",
		map[string]interface{}{
			"Endpoint": cmd.Config.Target(),
		})
	cmd.UI.DisplayText("Authenticating...")

	credentials := make(map[string]string)
	grantType := constant.GrantTypePassword
	if cmd.ClientCredentials {
		grantType = constant.GrantTypeClientCredentials
		credentials["client_id"] = username
		credentials["client_secret"] = password
	} else {
		credentials = map[string]string{
			"username": username,
			"password": password,
		}
	}

	err = cmd.Actor.Authenticate(credentials, cmd.Origin, grantType)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()
	cmd.UI.DisplayTextWithFlavor(
		"Use '{{.Command}}' to view or set your target org and space.",
		map[string]interface{}{
			"Command": fmt.Sprintf("%s target", cmd.Config.BinaryName()),
		})

	return nil
}

func (cmd AuthCommand) getUsernamePassword() (string, string, error) {
	var (
		userMissing     bool
		passwordMissing bool
	)

	username := cmd.RequiredArgs.Username
	if username == "" {
		if envUser := cmd.Config.CFUsername(); envUser != "" {
			username = envUser
		} else {
			userMissing = true
		}
	}

	password := cmd.RequiredArgs.Password
	if password == "" {
		if envPassword := cmd.Config.CFPassword(); envPassword != "" {
			password = envPassword
		} else {
			passwordMissing = true
		}
	}

	if userMissing || passwordMissing {
		return "", "", translatableerror.MissingCredentialsError{
			MissingUsername: userMissing,
			MissingPassword: passwordMissing,
		}
	}

	return username, password, nil
}
