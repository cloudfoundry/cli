package v2

import (
	"os"

	"code.cloudfoundry.org/cli/actor/v2action"
	oldCmd "code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v2/shared"
)

//go:generate counterfeiter . CreateUserActor

type CreateUserActor interface {
	NewUser(username string, password string) (v2action.User, v2action.Warnings, error)
}

type CreateUserCommand struct {
	RequiredArgs    flag.Authentication `positional-args:"yes"`
	usage           interface{}         `usage:"CF_NAME create-user USERNAME PASSWORD"`
	relatedCommands interface{}         `related_commands:"passwd, set-org-role, set-space-role"`

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
	if cmd.Config.Experimental() == false {
		oldCmd.Main(os.Getenv("CF_TRACE"), os.Args)
		return nil
	}

	cmd.UI.DisplayText(command.ExperimentalWarning)
	cmd.UI.DisplayNewline()

	err := command.CheckTarget(cmd.Config, false, false)
	if err != nil {
		return err
	}

	cmd.UI.DisplayText("Creating user {{.TargetUser}}...", map[string]interface{}{
		"TargetUser": cmd.RequiredArgs.Username,
	})

	_, warnings, err := cmd.Actor.NewUser(cmd.RequiredArgs.Username, cmd.RequiredArgs.Password)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()
	cmd.UI.DisplayNewline()
	cmd.UI.DisplayText("TIP: Assign roles with '{{.BinaryName}} set-org-role' and '{{.BinaryName}} set-space-role'.", map[string]interface{}{
		"BinaryName": cmd.Config.BinaryName(),
	})

	return nil
}
