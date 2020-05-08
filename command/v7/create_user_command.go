package v7

import (
	"strings"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/api/uaa"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type CreateUserCommand struct {
	command.BaseCommand

	Args            flag.CreateUser `positional-args:"yes"`
	Origin          string          `long:"origin" description:"Origin for mapping a user account to a user in an external identity provider"`
	PasswordPrompt  bool            `long:"password-prompt" description:"Prompt interactively for password"`
	usage           interface{}     `usage:"CF_NAME create-user USERNAME PASSWORD\n   CF_NAME create-user USERNAME [--origin ORIGIN]\n   CF_NAME create-user USERNAME --password-prompt\n\nEXAMPLES:\n   cf create-user j.smith@example.com S3cr3t                  # internal user\n   cf create-user j.smith@example.com --origin ldap           # LDAP user\n   cf create-user j.smith@example.com --origin provider-alias # SAML or OpenID Connect federated user"`
	relatedCommands interface{}     `related_commands:"passwd, set-org-role, set-space-role"`
}

func (cmd *CreateUserCommand) Execute(args []string) error {
	var password string
	var err error

	if cmd.passwordRequired() {
		return translatableerror.RequiredArgumentError{
			ArgumentName: "PASSWORD",
		}
	}

	//	Is the admin logged in?
	err = cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	origin := cmd.Origin
	if cmd.Origin == "" {
		origin = constant.DefaultOriginUaa
	}

	//	Does the new user already exist?
	_, err = cmd.Actor.GetUser(cmd.Args.Username, origin)
	if err == nil {
		// User already exists
		cmd.UI.DisplayTextWithFlavor("Creating user {{.TargetUser}}...", map[string]interface{}{
			"TargetUser": cmd.Args.Username,
		})

		cmd.UI.DisplayWarning("User '{{.User}}' already exists.", map[string]interface{}{
			"User": cmd.Args.Username,
		})

		cmd.UI.DisplayOK()
		return nil
	}

	//	Get the password
	if cmd.Args.Password != nil {
		password = *cmd.Args.Password
	}

	if cmd.PasswordPrompt {
		password, err = cmd.UI.DisplayPasswordPrompt("Password")
		if err != nil {
			return err
		}
	}

	//	Create the user!
	cmd.UI.DisplayTextWithFlavor("Creating user {{.TargetUser}}...", map[string]interface{}{
		"TargetUser": cmd.Args.Username,
	})

	_, warnings, err := cmd.Actor.CreateUser(cmd.Args.Username, password, cmd.Origin)
	cmd.UI.DisplayWarnings(warnings)

	if err != nil {
		if _, ok := err.(uaa.ConflictError); ok {
			cmd.UI.DisplayWarning("User '{{.User}}' already exists.", map[string]interface{}{
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
	if (cmd.Origin == "" || strings.ToLower(cmd.Origin) == constant.DefaultOriginUaa) && !cmd.PasswordPrompt {
		if cmd.Args.Password == nil {
			return true
		}
	}
	return false
}
