package v6

import (
	"fmt"

	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/api/uaa/constant"
	"code.cloudfoundry.org/cli/api/uaa/uaaversion"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v6/shared"
)

//go:generate counterfeiter . AuthActor

type AuthActor interface {
	Authenticate(ID string, secret string, origin string, grantType constant.GrantType) error
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

	ccClient, uaaClient, err := shared.NewClients(config, ui, true)
	if err != nil {
		return err
	}
	cmd.Actor = v2action.NewActor(ccClient, uaaClient, config)

	return nil
}

func (cmd AuthCommand) Execute(args []string) error {
	if len(cmd.Origin) > 0 {
		err := command.MinimumUAAAPIVersionCheck(cmd.Actor.UAAAPIVersion(), uaaversion.MinVersionOrigin, "Option '--origin'")
		if err != nil {
			return err
		}
	}

	if cmd.ClientCredentials && cmd.Origin != "" {
		return translatableerror.ArgumentCombinationError{
			Args: []string{"--client-credentials", "--origin"},
		}
	}

	username, password, err := cmd.checkEnvVariables()
	if err != nil {
		return err
	}

	err = command.WarnIfCLIVersionBelowAPIDefinedMinimum(cmd.Config, cmd.Actor.CloudControllerAPIVersion(), cmd.UI)
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor(
		"API endpoint: {{.Endpoint}}",
		map[string]interface{}{
			"Endpoint": cmd.Config.Target(),
		})
	cmd.UI.DisplayText("Authenticating...")

	grantType := constant.GrantTypePassword
	if cmd.ClientCredentials {
		grantType = constant.GrantTypeClientCredentials
	}

	err = cmd.Actor.Authenticate(username, password, cmd.Origin, grantType)
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

func (cmd AuthCommand) checkEnvVariables() (string, string, error) {
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
