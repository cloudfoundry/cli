package v7

import (
	"strings"

	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/clock"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/api/uaa"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v7/shared"
)

//go:generate counterfeiter . CreateUserActor

type CreateUserActor interface {
	CreateUser(username string, password string, origin string) (v7action.User, v7action.Warnings, error)
}

type CreateUserCommand struct {
	Args            flag.CreateUser `positional-args:"yes"`
	Origin          string          `long:"origin" description:"Origin for mapping a user account to a user in an external identity provider"`
	PasswordPrompt  bool            `long:"password-prompt" description:"Prompt interactively for password"`
	usage           interface{}     `usage:"CF_NAME create-user USERNAME PASSWORD\n   CF_NAME create-user USERNAME --origin ORIGIN\n   CF_NAME create-user USERNAME --password-prompt\n\nEXAMPLES:\n   cf create-user j.smith@example.com S3cr3t                  # internal user\n   cf create-user j.smith@example.com --origin ldap           # LDAP user\n   cf create-user j.smith@example.com --origin provider-alias # SAML or OpenID Connect federated user"`
	relatedCommands interface{}     `related_commands:"passwd, set-org-role, set-space-role"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       CreateUserActor
}

func (cmd *CreateUserCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	sharedActor := sharedaction.NewActor(config)
	cmd.SharedActor = sharedActor

	ccClient, uaaClient, err := shared.NewClients(config, ui, true, "")
	if err != nil {
		return err
	}
	cmd.Actor = v7action.NewActor(ccClient, config, sharedActor, uaaClient, clock.NewClock())
	return nil
}

func (cmd *CreateUserCommand) Execute(args []string) error {
	// cmd.Args.Password is intentionally set to a pointer such that we can check
	// if it is passed (otherwise we can't differentiate between the default
	// empty string and a passed in empty string.
	var password string
	var err error

	if cmd.passwordRequired() {
		return translatableerror.RequiredArgumentError{
			ArgumentName: "PASSWORD",
		}
	}

	if cmd.Args.Password != nil {
		password = *cmd.Args.Password
	} else {
		password = ""
		if cmd.PasswordPrompt {
			password, err = cmd.UI.DisplayPasswordPrompt("Password")
			if err != nil {
				return err
			}
		}
	}

	err = cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Creating user {{.TargetUser}}...", map[string]interface{}{
		"TargetUser": cmd.Args.Username,
	})

	_, warnings, err := cmd.Actor.CreateUser(cmd.Args.Username, password, cmd.Origin)
	cmd.UI.DisplayWarnings(warnings)

	if err != nil {
		if _, ok := err.(uaa.ConflictError); ok {
			cmd.UI.DisplayWarningV7("User '{{.User}}' already exists.", map[string]interface{}{
				"User": cmd.Args.Username,
			})
			cmd.UI.DisplayOK()
			return nil
		} else {
			return err
		}
	}

	cmd.UI.DisplayOK()
	cmd.UI.DisplayText("TIP: Assign roles with '{{.BinaryName}} set-org-role' and '{{.BinaryName}} set-space-role'.", map[string]interface{}{
		"BinaryName": cmd.Config.BinaryName(),
	})

	return nil
}

func (cmd *CreateUserCommand) passwordRequired() bool {
	if (cmd.Origin == "" || strings.ToLower(cmd.Origin) == "uaa") && !cmd.PasswordPrompt {
		if cmd.Args.Password == nil {
			return true
		}
	}
	return false
}
