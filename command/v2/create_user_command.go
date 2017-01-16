package v2

import (
	"strings"

	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/api/uaa"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v2/shared"
)

//go:generate counterfeiter . CreateUserActor

type CreateUserActor interface {
	NewUser(username string, password string, origin string) (v2action.User, v2action.Warnings, error)
}

type CreateUserCommand struct {
	Args            flag.CreateUser `positional-args:"yes"`
	Origin          string          `long:"origin" description:"Origin for mapping a user account to a user in an external identity provider"`
	usage           interface{}     `usage:"CF_NAME create-user USERNAME PASSWORD\n   CF_NAME create-user USERNAME --origin ORIGIN\n\nEXAMPLES:\n   cf create-user j.smith@example.com S3cr3t                  # internal user\n   cf create-user j.smith@example.com --origin ldap           # LDAP user\n   cf create-user j.smith@example.com --origin provider-alias # SAML or OpenID Connect federated user"`
	relatedCommands interface{}     `related_commands:"passwd, set-org-role, set-space-role"`

	UI     command.UI
	Config command.Config
	Actor  CreateUserActor
}

func (cmd *CreateUserCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config

	ccClient, uaaClient, err := shared.NewClients(config, ui)
	if err != nil {
		return err
	}
	cmd.Actor = v2action.NewActor(ccClient, uaaClient)

	return nil
}

func (cmd *CreateUserCommand) Execute(args []string) error {
	// cmd.Args.Password is intentionally set to a pointer such that we can check
	// if it is passed (otherwise we can't differentiate between the default
	// empty string and a passed in empty string.
	var password string

	if (cmd.Origin == "" || strings.ToLower(cmd.Origin) == "uaa") && cmd.Args.Password == nil {
		return command.RequiredArgumentError{
			ArgumentName: "PASSWORD",
		}
	}

	if cmd.Args.Password != nil {
		password = *cmd.Args.Password
	} else {
		password = ""
	}

	err := command.CheckTarget(cmd.Config, false, false)
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Creating user {{.TargetUser}}...", map[string]interface{}{
		"TargetUser": cmd.Args.Username,
	})

	_, warnings, err := cmd.Actor.NewUser(cmd.Args.Username, password, cmd.Origin)
	cmd.UI.DisplayWarnings(warnings)

	if err != nil {
		if _, ok := err.(uaa.ConflictError); ok {
			cmd.UI.DisplayWarning("user {{.User}} already exists", map[string]interface{}{
				"User": cmd.Args.Username,
			})
		} else {
			cmd.UI.DisplayTextWithFlavor("Error creating user {{.User}}.", map[string]interface{}{
				"User": cmd.Args.Username,
			})
			return err
		}
	}

	cmd.UI.DisplayOK()
	cmd.UI.DisplayNewline()
	cmd.UI.DisplayText("TIP: Assign roles with '{{.BinaryName}} set-org-role' and '{{.BinaryName}} set-space-role'.", map[string]interface{}{
		"BinaryName": cmd.Config.BinaryName(),
	})

	return nil
}
